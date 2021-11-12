package workerpool

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

func worker(ctx context.Context, jobs <-chan *Job, results chan<- Result) error {

	for {
		select {
		case job, ok := <-jobs:
			if !ok {
				return nil
			}
			// fan-in job execution multiplexing results into the results channel
			results <- job.Execute(ctx)

		case <-ctx.Done():
			fmt.Printf("cancelled worker. Error detail: %v\n", ctx.Err())
			results <- Result{
				Err: ctx.Err(),
			}
			return ctx.Err()
		}
	}
}

type WorkerPool struct {
	workersCount int
	jobs         chan *Job
	results      chan Result
	Done         chan struct{}
}

func NewWorkerPool(wcount int) WorkerPool {
	return WorkerPool{
		workersCount: wcount,
		jobs:         make(chan *Job, wcount),
		results:      make(chan Result, wcount),
		Done:         make(chan struct{}),
	}
}

func (wp WorkerPool) Run(ctx context.Context) {
	var eg errgroup.Group

	for i := 0; i < wp.workersCount; i++ {
		eg.Go(func() error {
			// fan out worker goroutines
			//reading from jobs channel and
			//pushing calcs into results channel
			err := worker(ctx, wp.jobs, wp.results)
			return err
		})

	}

	go func() {
		if err := eg.Wait(); err != nil {
			fmt.Printf("Error on worker %v", err)
		}

		close(wp.Done)
		close(wp.results)
	}()

}

func (wp WorkerPool) Results() <-chan Result {
	return wp.results
}

func (wp WorkerPool) GenerateFrom(jobsBulk []*Job) {
	for i, _ := range jobsBulk {
		wp.jobs <- jobsBulk[i]
	}
	close(wp.jobs)
}
