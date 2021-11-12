package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"workerpool-pattern/workerpool"
)

type config struct {
	totalworker int
	totaltask   int
}

// type result struct {
// 	id    int
// 	value int
// }

func process(config config) <-chan bool {

	// fmt.Println("...start process")

	waitC := make(chan bool)

	ctx := context.Background()
	wp := workerpool.NewWorkerPool(ctx, config.totalworker)

	chan_result, chan_err := wp.Run()
	go func() {

		defer wp.CloseQueue()

		for i := 0; i < config.totaltask; i++ {
			id := i + 1
			switch id % 3 {
			case 0:
				wp.AddTask(workerpool.Task{"PRINT", 0, print1})
			case 1:
				wp.AddTask(workerpool.Task{"SCAN", 0, scan1})
			case 2:
				wp.AddTask(workerpool.Task{"EMAIL", 0, email1})
			}
		}

	}()

	go func() {
		for result := range chan_result {
			fmt.Println("Result: ", result.Item)
		}

	}()

	go func() {
		err := <-chan_err
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		waitC <- true
	}()

	return waitC
}

func main() {
	log.SetFlags(log.Ltime)

	config := config{totalworker: 3, totaltask: 15}

	// For monitoring purpose.

	// go func() {
	// 	for {
	// 		log.Printf("[main] Total current goroutine: %d", runtime.NumGoroutine())
	// 		// time.Sleep(1 * time.Second)
	// 	}
	// }()

	waitC := process(config)
	<-waitC
}

func print1() {
	fmt.Println(".... task: print")
}

func scan1() {
	fmt.Println(".... task: scan")
}

func email1() {
	fmt.Println(".... task: email")
}
