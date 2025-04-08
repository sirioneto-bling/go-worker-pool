package work

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// Executor defines the interface for a job that can be executed.
type Executor interface {
	// Execute performs the job logic and returns a result or an error.
	Execute() (any, error)
	// OnError handles any error returned by Execute.
	OnError(error)
	// ShouldReturnResult defines if the job should populate result channel
	ShouldReturnResult() bool
}

// Result represents the outcome of a job execution.
type Result struct {
	Value    any
	Error    error
	Executor Executor
}

// Pool manages a set of worker goroutines to execute jobs concurrently.
type Pool struct {
	numWorkers int
	jobs       chan Executor
	results    chan Result
	start      sync.Once
	stop       sync.Once
	jobWg      sync.WaitGroup
	done       chan struct{}
}

// NewPool initializes a new worker pool.
func NewPool(numWorkers, jobChannelSize int) (*Pool, error) {
	if numWorkers <= 0 {
		return nil, errors.New("numWorkers must be greater than zero")
	}
	if jobChannelSize < 0 {
		return nil, errors.New("jobChannelSize cannot be negative")
	}
	return &Pool{
		numWorkers: numWorkers,
		jobs:       make(chan Executor, jobChannelSize),
		results:    make(chan Result, jobChannelSize),
		done:       make(chan struct{}),
	}, nil
}

// Start spins up the worker goroutines.
func (p *Pool) Start(ctx context.Context) {
	p.start.Do(func() {
		for i := 0; i < p.numWorkers; i++ {
			go p.worker(ctx, i+1)
		}
	})
}

// Stop signals the pool to finish and waits for all jobs to complete.
func (p *Pool) Stop() {
	p.stop.Do(func() {
		p.WaitJobs()
		close(p.jobs)
		close(p.results)
		close(p.done)
	})
}

// AddJob adds a new job to the pool and tracks it in the WaitGroup.
func (p *Pool) AddJob(t Executor) {
	select {
	case <-p.done:
		return
	default:
		p.jobWg.Add(1)
		p.jobs <- t
	}
}

// WaitJobs block's until jobs wait group counter is zero
func (p *Pool) WaitJobs() {
	p.jobWg.Wait()
}

// Done block's until worker pool gracefull shutdown execution
func (p *Pool) Done() <-chan struct{} {
	return p.done
}

// Result returns worker job result channel
func (p *Pool) Result() <-chan Result {
	return p.results
}

// worker start jobs channel worker consumer
func (p *Pool) worker(ctx context.Context, workerNum int) {
	fmt.Printf("Worker %d started\n", workerNum)

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Worker %d stopping: context canceled\n", workerNum)
			return
		case <-p.Done():
			fmt.Printf("Worker %d stopping: pool stopped\n", workerNum)
			return
		case job, ok := <-p.jobs:
			if !ok {
				return
			}

			value, err := job.Execute()
			if err != nil {
				job.OnError(err)
			}

			if job.ShouldReturnResult() {
				p.results <- Result{
					Value:    value,
					Error:    err,
					Executor: job,
				}
			}

			p.jobWg.Done()
		}
	}
}
