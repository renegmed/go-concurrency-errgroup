package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

func JobWithCtx(ctx context.Context, jobID int) error {
	select {
	case <-ctx.Done():
		fmt.Printf("\ncontext cancelled job %v terminting\n", jobID)
		return nil
	case <-time.After(time.Second * time.Duration(rand.Intn(3))):
	}
	if rand.Intn(12) == jobID {
		fmt.Printf("Job %v failed.\n", jobID)
		return fmt.Errorf("job %v failed", jobID)
	}

	fmt.Printf("Job %v done.\n", jobID)
	return nil
}

func NewCtx() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sCh := make(chan os.Signal, 1)
		signal.Notify(sCh, syscall.SIGINT, syscall.SIGTERM)
		<-sCh
		cancel()
	}()
	return ctx
}

func main() {
	//eg, ctx := errgroup.WithContext(context.Background())
	eg, ctx := errgroup.WithContext(NewCtx())

	for i := 0; i < 10; i++ {
		jobID := i
		eg.Go(func() error {
			return JobWithCtx(ctx, jobID)
		})
	}

	if err := eg.Wait(); err != nil {
		fmt.Println("Encountered error:", err)
		return
	}
	fmt.Println("Successfully finished.")
}
