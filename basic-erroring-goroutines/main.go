/*

We are going to run 3 goroutines.

If any of them would delay more than 80 milliseconds to finish its own
job, we will force it to return an error because we are in a hurry. If
that happens, all the running goroutines will be terminated too.

Otherwise obviously all will work fine.

*/
package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/sync/errgroup"
)

func main() {
	fmt.Println(">> START")

	// Prevent picking up the same random number all the time for sleeping.
	rand.Seed(time.Now().UnixNano())

	goErrGroup, ctx := errgroup.WithContext(context.Background())

	goErrGroup.Go(func() error {
		return doSomething(ctx, 1)
	})
	goErrGroup.Go(func() error {
		return doSomething(ctx, 2)
	})
	goErrGroup.Go(func() error {
		return doSomething(ctx, 3)
	})
	goErrGroup.Go(func() error {
		return doSomething(ctx, 4)
	})

	// Wait for the first error from any goroutine.
	if err := goErrGroup.Wait(); err != nil {
		fmt.Println(err)
	}

	fmt.Println(">> FINISH")
}

func doSomething(ctx context.Context, id int) error {
	fmt.Printf("START: goroutine %d\n", id)

	// Pick a random number to simulate time it takes to finish the job.
	delay := rand.Intn(100)
	if delay > 80 {
		return fmt.Errorf("++++FAIL: goroutine %d: %dms", id, delay)
	}
	time.Sleep(time.Duration(delay) * time.Millisecond)

	fmt.Printf("END: goroutine %d\n", id)

	return nil
}
