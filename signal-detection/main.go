/*

Example for using go's sync.errgroup together with signal detection signal.Notify
to stop all running goroutines

*/
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

func main() {

	//ctx, cancel := context.WithCancel(context.Background())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5) //time.Millisecond*100)
	g, gctx := errgroup.WithContext(ctx)

	// goroutine to check for signals to gracefully finish all functions
	g.Go(func() error {
		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

		select {
		case sig := <-signalChannel:
			fmt.Printf("\nReceived signal: %s\n", sig)
			cancel()
		case <-gctx.Done():
			fmt.Printf("done, closing signal goroutine\n")
			return gctx.Err()
		}

		return nil
	})

	// just a ticker every 2s
	g.Go(func() error {
		ticker := time.NewTicker(2 * time.Second)
		for {
			select {
			case <-ticker.C:
				fmt.Printf("ticker 2s ticked\n")
				// testcase what happens if an error occured
				//return fmt.Errorf("test error ticker 2s")
			case <-gctx.Done():
				fmt.Printf("closing ticker 2s goroutine\n")
				return gctx.Err()
			}
		}
	})

	// just a ticker every 1s
	g.Go(func() error {
		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ticker.C:
				fmt.Printf("ticker 1s ticked\n")
			case <-gctx.Done():
				fmt.Printf("closing ticker 1s goroutine\n")
				return gctx.Err()
			}
		}
	})

	// force a stop after 60s
	// time.AfterFunc(10*time.Second, func() {
	// 	fmt.Printf("force finished after 10s\n")
	// 	cancel()
	// })

	// wait for all errgroup goroutines
	// err := g.Wait()
	// if err != nil {
	// 	if errors.Is(err, context.Canceled) {
	// 		fmt.Print("context was canceled")
	// 	} else {
	// 		fmt.Printf("received error: %v", err)
	// 	}
	// } else {
	// 	fmt.Println("finished clean")
	// }

	if err := g.Wait(); err == nil || err == context.Canceled {
		fmt.Println("...finished clean")
	} else {
		fmt.Printf("...received error: %v\n", err)
	}

}
