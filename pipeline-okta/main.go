package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"pipeline-design-pattern/pipeline"
)

func main() {

	var choice int

	for {
		fmt.Println("Which app to use?")
		fmt.Print("Enter 1=serial 2=parallel 3=exit:")

		n, err := fmt.Scanf("%d", &choice)
		if n != 1 || err != nil {
			fmt.Println("Follow directions!")
			return
		}

		switch choice {
		case 1:
			serial()

		case 2:
			parallel()
		case 3:
			os.Exit(1)
		}

	}

	N := 5
	startTime := time.Now()

	for i := 0; i < N; i++ {
		result := addFoo(addQuoute(square(multiplyTwo(i))))
		fmt.Printf("Result: %s\n", result)
	}

	fmt.Printf("Elapsed time without concurrency: %s", time.Since(startTime)) // ~40 seconds

	outC := pipeline.New(func(inC chan interface{}) {
		defer close(inC)
		for i := 0; i < N; i++ {
			inC <- i
		}
	}).
		Pipe(func(in interface{}) (interface{}, error) {
			return multiplyTwo(in.(int)), nil
		}).
		Pipe(func(in interface{}) (interface{}, error) {
			return square(in.(int)), nil
		}).
		Pipe(func(in interface{}) (interface{}, error) {
			return addQuoute(in.(int)), nil
		}).
		Pipe(func(in interface{}) (interface{}, error) {
			return addFoo(in.(string)), nil
		}).
		Merge()

	startTimeC := time.Now()
	for result := range outC {
		fmt.Printf("Result: %s\n", result)
	}

	fmt.Printf("Elapsed time with concurrency: %s", time.Since(startTimeC)) // ~16 seconds
}

func serial() {
	N := 5
	startTime := time.Now()

	for i := 0; i < N; i++ {
		result := addFoo(addQuoute(square(multiplyTwo(i))))
		fmt.Printf("Result: %s\n", result)
	}

	fmt.Printf("Elapsed time without concurrency: %s\n", time.Since(startTime)) // ~40 seconds

}

func parallel() {
	N := 5

	outC := pipeline.New(context.Background(), func(inC chan interface{}) {
		defer close(inC)
		for i := 0; i < N; i++ {
			inC <- i
		}
	}).
		Pipe(func(in interface{}) (interface{}, error) {
			return multiplyTwo(in.(int)), nil
		}).
		Pipe(func(in interface{}) (interface{}, error) {
			return square(in.(int)), nil
		}).
		Pipe(func(in interface{}) (interface{}, error) {
			return addQuoute(in.(int)), nil
		}).
		Pipe(func(in interface{}) (interface{}, error) {
			return addFoo(in.(string)), nil
		}).
		Merge()

	startTimeC := time.Now()
	for result := range outC {
		fmt.Printf("Result: %s\n", result)
	}

	fmt.Printf("Elapsed time with concurrency: %s\n", time.Since(startTimeC)) // ~16 seconds
}

func multiplyTwo(v int) int {
	//time.Sleep(2 * time.Second)
	return v * 2
}

func square(v int) int {
	//time.Sleep(2 * time.Second)
	return v * v
}

func addQuoute(v int) string {
	//time.Sleep(2 * time.Second)
	return fmt.Sprintf("'%d'", v)
}

func addFoo(v string) string {
	//time.Sleep(2 * time.Second)
	return fmt.Sprintf("%s - Foo", v)
}
