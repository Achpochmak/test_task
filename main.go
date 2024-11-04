package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
)

type Worker struct {
	ID   int
	Quit chan bool
}

type WorkerPool struct {
	Workers []*Worker
	JobChan chan string
	wg      sync.WaitGroup
}

func NewWorker(id int) *Worker {
	return &Worker{
		ID:   id,
		Quit: make(chan bool),
	}
}

func (w *Worker) Start(wg *sync.WaitGroup, jobChan chan string) {
	defer wg.Done()
	for {
		select {
		case job := <-jobChan:
			fmt.Printf("Worker %d: Job: %s\n", w.ID, job)
		case <-w.Quit:
			fmt.Printf("Worker %d: Quiting\n", w.ID)
			return
		}
	}
}

func NewWorkerPool(numWorkers int) *WorkerPool {
	pool := &WorkerPool{
		JobChan: make(chan string, 10),
	}

	for i := 0; i < numWorkers; i++ {
		worker := NewWorker(i + 1)
		pool.Workers = append(pool.Workers, worker)
		pool.wg.Add(1)
		go worker.Start(&pool.wg, pool.JobChan)
	}

	return pool
}

func (p *WorkerPool) AddWorker() {
	worker := NewWorker(len(p.Workers) + 1)
	p.Workers = append(p.Workers, worker)
	p.wg.Add(1)
	go worker.Start(&p.wg, p.JobChan)
	fmt.Printf("Added worker, current count: %d\n", len(p.Workers))
}

func (p *WorkerPool) RemoveWorker() {
	lenWorkers := len(p.Workers)
	if lenWorkers == 0 {
		fmt.Println("Worker pool is already empty")
		return
	}

	worker := p.Workers[lenWorkers-1]
	worker.Quit <- true
	p.Workers = p.Workers[:lenWorkers-1]
	fmt.Printf("Removed worker, current count: %d\n", len(p.Workers))
}

func (p *WorkerPool) Shutdown() {
	for _, worker := range p.Workers {
		worker.Quit <- true
	}

	p.wg.Wait()
	close(p.JobChan)
}

func main() {
	runtime.GOMAXPROCS(0)
	pool := NewWorkerPool(3)
	reader := bufio.NewReader(os.Stdin)

	for {
		input, _ := reader.ReadString('\n')
		cmd := strings.TrimSpace(input)
		switch {
		case cmd == "quit":
			fmt.Println("Shutting down...")
			pool.Shutdown()
			return
		case cmd == "add":
			pool.AddWorker()
		case cmd == "remove":
			pool.RemoveWorker()
		default:
			pool.JobChan <- cmd
		}
	}
}
