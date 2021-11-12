package workerpool

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

// WorkerPool is a contract for Worker Pool implementation
type WorkerPool interface {
	Run() (<-chan Result, <-chan error)
	//AddTask(task func())
	AddTask(task Task)
	CloseQueue()
}

type workerPool struct {
	maxWorker   int
	queuedTaskC chan Task
	context     context.Context
}

type Task struct {
	Name        string
	PerformedBy int
	Job         func()
}

type Result struct {
	Item string
}

// NewWorkerPool will create an instance of WorkerPool.
func NewWorkerPool(ctx context.Context, maxWorker int) WorkerPool {
	wp := &workerPool{
		context:     ctx,
		maxWorker:   maxWorker,
		queuedTaskC: make(chan Task),
	}

	return wp
}

func (wp *workerPool) Run() (<-chan Result, <-chan error) {
	return wp.run()
}

func (wp *workerPool) AddTask(task Task) {
	wp.queuedTaskC <- task
}

func (wp *workerPool) CloseQueue() {
	close(wp.queuedTaskC)
}

func (wp *workerPool) GetTotalQueuedTask() int {
	return len(wp.queuedTaskC)
}

func (wp *workerPool) run() (<-chan Result, <-chan error) {
	ch := make(chan error)
	resch := make(chan Result)

	g, ctx := errgroup.WithContext(wp.context)

	for i := 0; i < wp.maxWorker; i++ {
		wID := i + 1
		//log.Printf("[WorkerPool] Worker %d has been spawned", wID)

		g.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					fmt.Printf("...worker %d error: %v\n", wID, ctx.Err())
					return ctx.Err()
				case task, ok := <-wp.queuedTaskC:
					if !ok {
						fmt.Printf("... no more work to do for %d\n", wID)
						return nil
					}
					task.Job()
					resch <- Result{fmt.Sprintf("task '%s' completed by %d", task.Name, wID)}
				}
			}
		})

		// func(workerID int) {
		// 	for task := range wp.queuedTaskC {
		// 		log.Printf("[WorkerPool] Worker %d start processing task", wID)
		// 		task()
		// 		log.Printf("[WorkerPool] Worker %d finish processing task", wID)
		// 	}
		// }(wID)
	}
	go func() {
		err := g.Wait()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}

		ch <- err
		close(resch)

	}()

	return resch, ch
}
