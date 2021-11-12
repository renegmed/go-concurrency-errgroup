package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

type workItem struct {
	work int
}

type counter struct {
	toprocess int
	processed int
}

var mux sync.Mutex
var cntr counter

func work(ctx context.Context, workItemCount int, workers int) chan counter {
	cntr = counter{}
	ch := make(chan counter)
	// Create a channel with buffer = workItemCount
	workChan := make(chan workItem, workItemCount)

	//var wg sync.WaitGroup
	g, ctx := errgroup.WithContext(ctx)

	// Populate the work items
	go func() {
		defer close(workChan) // when closed, it's a signal that no more items to send
		for i := 0; i < workItemCount; i++ {
			workChan <- workItem{i}
			mux.Lock()
			cntr.toprocess += 1
			mux.Unlock()
		}
	}()

	// Launch workers
	//log.Println("Launching workers...")

	for i := 0; i < workers; i++ {
		worker := i

		g.Go(func() error {

			//log.Printf("[Worker %d] Started", worker)
			for {
				select {
				case <-ctx.Done():
					fmt.Printf("Context completed %v\n", ctx.Err())
					return ctx.Err()
				case workItem, ok := <-workChan:

					//log.Printf("....[Worker %d] HAS work to do", worker)

					if !ok {
						log.Printf("....[Worker %d] No more work to do", worker)
						return nil
					}
					// induce an error
					if workItem.work == 101 {
						return fmt.Errorf("illegal item %d", workItem)
					}

					// We've got work
					doWork(worker, workItem)

				}
			}
		})

	}

	// Close the channel and wait for all waiters to finish
	go func() {
		err := g.Wait()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}

		ch <- cntr

	}()

	return ch
}

func doWork(worker int, work workItem) {
	//log.Printf("[Worker %d] Working on %d\n", worker, work.work)
	// Simulate work
	time.Sleep(10 * time.Millisecond)
	//log.Printf("[Worker %d] Completed %d\n", worker, work.work)
	mux.Lock()
	cntr.processed += 1
	mux.Unlock()
}

func main() {
	var ctx context.Context

	var cancel context.CancelFunc

	var choice int
	for {
		fmt.Println("Which app to use?")
		fmt.Print("Enter 1=no timetout 2=no timeout with error 3=with timeout 4=exit:")

		n, err := fmt.Scanf("%d", &choice)
		if n != 1 || err != nil {
			fmt.Println("Follow directions!")
			return
		}

		var numOfItems = 100
		switch choice {
		case 1:
			ctx = context.Background()
		case 2:
			ctx = context.Background()
			numOfItems = 110
		case 3:
			ctx, cancel = context.WithTimeout(context.Background(), 1*time.Millisecond)
			defer cancel()
		case 4:
			os.Exit(1)
		}

		wait := work(ctx, numOfItems, 10)

		fmt.Println("...wait")

		c := <-wait

		fmt.Println("----------------", c)
	}

}
