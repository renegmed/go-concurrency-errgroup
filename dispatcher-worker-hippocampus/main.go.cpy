/*

immediately cancels the other jobs when an error occurs in any goroutine

*/
package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	numWorkers      = 10
	queueLength     = 100
	numJobs         = 100
	maxWaitMilliSec = 100 //1000
	thres           = 90  //990
)

type Job struct {
	id           int
	waitMilliSec int
}

type JobError struct {
	id      int
	message string
}

type Dispatcher struct {
	queue  chan *Job
	errors chan *JobError
	eg     *errgroup.Group
}

func NewDispatcher(eg *errgroup.Group) *Dispatcher {
	return &Dispatcher{
		queue:  make(chan *Job, queueLength), // buffered channel
		errors: make(chan *JobError),
		eg:     eg,
	}
}

func (d *Dispatcher) StartDispatchingWork(ctx context.Context) { // start receiving job and assign to worker
	for i := 0; i < numWorkers; i++ {
		d.eg.Go(func() error {
			for j := range d.queue { // receive jobs from channel
				err := worker_do_work(ctx, j)
				if err != nil {
					d.Error(j.id, err)
					return nil
				}
			}
			return nil
		})
	}
}

func (d *Dispatcher) Append(job *Job) { // send job to a worker
	d.queue <- job
}

func (d *Dispatcher) Error(id int, err error) { // send job to a worker
	d.errors <- &JobError{id, fmt.Sprintf("%v", err)}
}

func worker_do_work(ctx context.Context, job *Job) error {
	select {
	case <-ctx.Done():
		fmt.Printf("Canceled the job #%d\n", job.id)
		return nil
	default:
		fmt.Printf("Working on the job #%d. Wait for %d ms.\n", job.id, job.waitMilliSec)
		if job.waitMilliSec > thres {
			//fmt.Printf("cannot wait for more than %d ms: job #%d; %d ms\n", thres, job.id, job.waitMilliSec)
			//return nil
			return fmt.Errorf("cannot wait for more than %d ms: job #%d; %d ms", thres, job.id, job.waitMilliSec)
			//fmt.Printf("cannot wait for more than %d ms: job #%d; %d ms", thres, job.id, job.waitMilliSec)
			//return nil
		}
		time.Sleep(time.Duration(job.waitMilliSec) * time.Millisecond)
		return nil
	}
}

func main() {

	rand.Seed(time.Now().UnixNano())

	eg, ctx := errgroup.WithContext(context.Background())
	d := NewDispatcher(eg)

	d.StartDispatchingWork(ctx)

	// create and dispatch jobs
	for i := 0; i < numJobs; i++ { // numJobs: 100
		milliSec := rand.Intn(maxWaitMilliSec)
		d.Append(&Job{
			id:           i,
			waitMilliSec: milliSec,
		})
	}

	go func() {
		for err := range d.errors {
			fmt.Printf("...Errors: %d, %s\n", err.id, err.message)
		}
		close(d.errors)
	}()

	close(d.queue)

	err := d.eg.Wait()
	if err != nil {
		fmt.Println("...error on wait():", err)
	}

}
