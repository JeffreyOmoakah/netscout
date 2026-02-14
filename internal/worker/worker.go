package worker

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/JeffreyOmoakah/netscout.git/internal/result"
)

// Task represents a single scan task
type Task struct {
	IP   string
	Port int
}

// Worker performs TCP port scanning
type Worker struct {
	id         int
	taskChan   <-chan Task
	resultChan chan<- *result.Result
	timeout    time.Duration
}

// NewWorker creates a new worker
func NewWorker(id int, taskChan <-chan Task, resultChan chan<- *result.Result, timeout time.Duration) *Worker {
	return &Worker{
		id:         id,
		taskChan:   taskChan,
		resultChan: resultChan,
		timeout:    timeout,
	}
}

// Start begins the worker's task processing loop
func (w *Worker) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-w.taskChan:
			if !ok {
				return
			}
			w.scan(task)
		}
	}
}

// scan performs a TCP connect scan on a single target
func (w *Worker) scan(task Task) {
	startTime := time.Now()
	
	r := &result.Result{
		IP:        task.IP,
		Port:      task.Port,
		Timestamp: startTime,
	}

	address := net.JoinHostPort(task.IP, strconv.Itoa(task.Port))

	
	// Attempt TCP connection with timeout
	conn, err := net.DialTimeout("tcp", address, w.timeout)
	
	r.Duration = time.Since(startTime)

	if err != nil {
		// Determine if port is filtered or closed based on error type
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			r.Status = result.StatusFiltered
		} else {
			r.Status = result.StatusClosed
		}
		r.Error = err.Error()
	} else {
		r.Status = result.StatusOpen
		conn.Close()
	}

	w.resultChan <- r
}

// Pool manages a pool of workers
type Pool struct {
	workers    []*Worker
	taskChan   chan Task
	resultChan chan<- *result.Result
	timeout    time.Duration
	size       int
}

// NewPool creates a new worker pool
func NewPool(size int, resultChan chan<- *result.Result, timeout time.Duration) *Pool {
	// Buffer size should be large enough to hold many tasks
	// but not so large it consumes too much memory
	bufferSize := size * 10
	if bufferSize > 10000 {
		bufferSize = 10000
	}

	return &Pool{
		workers:    make([]*Worker, 0, size),
		taskChan:   make(chan Task, bufferSize),
		resultChan: resultChan,
		timeout:    timeout,
		size:       size,
	}
}

// Start initializes and starts all workers in the pool
func (p *Pool) Start(ctx context.Context) {
	for i := 0; i < p.size; i++ {
		worker := NewWorker(i, p.taskChan, p.resultChan, p.timeout)
		p.workers = append(p.workers, worker)
		go worker.Start(ctx)
	}
}

// Submit submits a task to the worker pool
func (p *Pool) Submit(ctx context.Context, task Task) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case p.taskChan <- task:
		return nil
	}
}

// Close closes the task channel and waits for all workers to finish
func (p *Pool) Close() {
	close(p.taskChan)
}

// GetTaskChannel returns the task channel (useful for direct access)
func (p *Pool) GetTaskChannel() chan<- Task {
	return p.taskChan
}