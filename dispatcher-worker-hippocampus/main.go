/*

immediately cancels the other jobs when an error occurs in any goroutine

*/
package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	numWorkers      = 5
	queueLength     = 100
	numJobs         = 100
	maxWaitMilliSec = 100 //1000
	thres           = 90  //990
)

type config struct {
	numWorkers      int
	queueLength     int
	numJobs         int
	maxWaitMilliSec int
	thres           int
}

type Job struct {
	id           int
	waitMilliSec int
}

type JobError struct {
	id      int
	message string
}

// type Worker struct {
// 	id int
// }

type Dispatcher struct {
	queue  chan *Job
	errors chan *JobError
	config *config
	eg     *errgroup.Group
	mux    sync.Mutex
}

func NewDispatcher(eg *errgroup.Group, c *config) *Dispatcher {
	return &Dispatcher{
		queue:  make(chan *Job, c.queueLength), // buffered channel
		errors: make(chan *JobError),
		eg:     eg,
		config: &config{
			numWorkers:      c.numWorkers,
			queueLength:     c.queueLength,
			numJobs:         c.numJobs,
			maxWaitMilliSec: c.maxWaitMilliSec,
			thres:           c.thres,
		},
	}
}

func (d *Dispatcher) StartDispatchingWork(ctx context.Context) { // start receiving job and assign to worker
	for i := 0; i < d.config.numWorkers; i++ {
		i := i
		d.eg.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					//fmt.Printf("Canceled the job #%d\n", job.id)
					fmt.Println("...done")
					return ctx.Err()
				case job, ok := <-d.queue:
					if !ok {
						fmt.Printf("+++ no more job for %d\n", i)
						return nil
					}
					fmt.Printf("Working on the job #%d. Wait for %d ms.\n", job.id, job.waitMilliSec)
					// if job.waitMilliSec > thres {
					// 	//fmt.Printf("cannot wait for more than %d ms: job #%d; %d ms\n", thres, job.id, job.waitMilliSec)
					// 	//return nil
					// 	return fmt.Errorf("cannot wait for more than %d ms: job #%d; %d ms", thres, job.id, job.waitMilliSec)
					// 	//fmt.Printf("cannot wait for more than %d ms: job #%d; %d ms", thres, job.id, job.waitMilliSec)
					// 	//return nil
					// }
					// time.Sleep(time.Duration(job.waitMilliSec) * time.Millisecond)

				}
			}
		})
	}
}

func (d *Dispatcher) Append(job *Job) { // send job to a worker
	d.mux.Lock()
	d.queue <- job
	d.mux.Unlock()
}

func (d *Dispatcher) Error(id int, err error) { // send job to a worker
	d.errors <- &JobError{id, fmt.Sprintf("%v", err)}
}

func work_dispatcher(jobs []*Job, c *config) {

	eg, ctx := errgroup.WithContext(context.Background())
	d := NewDispatcher(eg, c)

	d.StartDispatchingWork(ctx)

	go func() {
		// wait for error or until processing is done
		err := d.eg.Wait()
		if err != nil {
			fmt.Println("...error on wait():", err)
		}

		// close(d.queue)
		fmt.Println("... done")

	}()

	// create and dispatch jobs
	go func() {
		defer close(d.queue)
		for i := 0; i < len(jobs); i++ {
			milliSec := rand.Intn(maxWaitMilliSec)
			job := jobs[i]
			job.waitMilliSec = milliSec
			//fmt.Println("---- append job", job.id)
			d.Append(job)
		}
	}()

}

func generateJobs(numOfJobs int) []*Job {
	var jobs []*Job
	for i := 0; i < numOfJobs; i++ {
		job := Job{id: i}
		jobs = append(jobs, &job)
	}

	return jobs
}

func main() {

	rand.Seed(time.Now().UnixNano())
	jobs := generateJobs(numJobs)
	c := &config{
		numWorkers:      numWorkers,
		queueLength:     len(jobs),
		numJobs:         len(jobs),
		maxWaitMilliSec: maxWaitMilliSec,
		thres:           thres,
	}

	work_dispatcher(jobs, c)
}
