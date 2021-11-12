package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"os"
	"workerpool-pattern/workerpool"
)

var ErrDefault = errors.New("wrong type of argument")

func main() {
	var choice int

	for {
		fmt.Println("Which app to use?")
		fmt.Print("Enter 1=without workerpool 2=with workerpool 3=exit:")

		n, err := fmt.Scanf("%d", &choice)
		if n != 1 || err != nil {
			fmt.Println("Follow directions!")
			return
		}

		switch choice {
		case 1:
			ctx, cancel := context.WithTimeout(context.TODO(), 120*time.Second) //time.Nanosecond*10)
			defer cancel()
			without_workpool(ctx, 1000000)
		case 2:
			ctx, cancel := context.WithTimeout(context.TODO(), 120*time.Second) //time.Nanosecond*10)
			defer cancel()
			with_workpool(ctx, 1000000, 8)
		case 3:
			os.Exit(1)
		}
	}

}
func without_workpool(ctx context.Context, numOfJobs int) {
	jobs := generateJobProducts(numOfJobs)
	startTime := time.Now()
	for _, job := range jobs {
		result := job.Execute(ctx)
		if result.Err != nil {
			fmt.Println("Error:\n", result.Err)
			os.Exit(1)
		}

		switch result.Value.(type) {
		case string:
			fmt.Println("string: ", result.Value.(string))
		case int:
			fmt.Println("int: ", result.Value.(int))
		case float32:
			fmt.Printf("float32: %.4f\n", result.Value.(float32))
		case Product:
			fmt.Printf("Product: %v\n", result.Value.(Product))

		default:
			fmt.Println("result type undetermined.")
		}
	}

	fmt.Printf("... done, elapsed time: %v\n", time.Since(startTime))
}

func with_workpool(ctx context.Context, numOfJobs, numOfWorkers int) {
	//jobs := generateJobs(numOfJobs)

	jobs := generateJobProducts(numOfJobs)

	wp := workerpool.NewWorkerPool(numOfWorkers)

	go wp.GenerateFrom(jobs)

	startTime := time.Now()
	go wp.Run(ctx)

	for {
		select {
		case r, ok := <-wp.Results():
			if !ok {
				continue
			}

			// i, err := strconv.ParseInt(string(r.Descriptor.ID), 10, 64)
			// if err != nil {
			// 	fmt.Printf("unexpected error: %v\n", err)
			// 	continue
			// }

			switch r.Value.(type) {
			case int:
				fmt.Println(r.Value.(int))
				// val := r.Value.(int)
				// if val != int(i)*2 {
				// 	fmt.Printf("wrong value %v; expected %v\n", val, int(i)*2)
				// }
			case string:
				//val := r.Value.(string)
				fmt.Println(r.Value.(string))

			case float32:
				fmt.Println(r.Value.(float32))
			case Product:
				fmt.Println(r.Value.(Product))

			default:
				fmt.Println("Returned value undetermined")
			}

		case <-wp.Done:
			fmt.Printf("... done, elapsed time: %v\n", time.Since(startTime))
			return
		default:
		}
	}
}

func multiplyByTwo(ctx context.Context, args interface{}) (interface{}, error) {
	argVal, ok := args.(int)
	if !ok {
		return nil, ErrDefault //fmt.Errorf("Error: Invalid number.")
	}

	return argVal * 2, nil
}

func divideByThree(ctx context.Context, args interface{}) (interface{}, error) {
	argVal, ok := args.(int)
	if !ok {
		return nil, ErrDefault
	}
	return float32(argVal) / float32(3), nil
}

func convertToString(ctx context.Context, args interface{}) (interface{}, error) {
	argVal, ok := args.(int)
	if !ok {
		return nil, ErrDefault // fmt.Errorf("Error: Invalid number.")
	}

	return strconv.Itoa(argVal) + " - by golang app", nil
}

type Product struct {
	Base       float32
	Multiplier float32
	Price      float32
}

func computePrice(ctx context.Context, args interface{}) (interface{}, error) {
	product, ok := args.(Product)
	if !ok {
		return nil, ErrDefault
	}
	product.Price = product.Base * product.Multiplier

	return product, nil

}
func execFn(i int) func(ctx context.Context, args interface{}) (interface{}, error) {
	switch i % 3 {
	case 0:
		return multiplyByTwo
	case 1:
		return divideByThree
	default:
		return convertToString
	}
}

func generateJobs(numOfJobs int) []*workerpool.Job {

	rand.Seed(time.Now().Unix())

	jobs := make([]*workerpool.Job, numOfJobs)
	for i := 0; i < numOfJobs; i++ {
		j := rand.Intn(500) + 25

		jobs[i] = &workerpool.Job{
			Descriptor: workerpool.JobDescriptor{
				ID:       workerpool.JobID(fmt.Sprintf("%v", i)),
				JType:    "anyType",
				Metadata: nil,
			},
			ExecFn: execFn(j),
			Args:   j,
		}
	}
	return jobs
}

func generateJobProducts(numOfJobs int) []*workerpool.Job {

	rand.Seed(time.Now().Unix())

	jobs := make([]*workerpool.Job, numOfJobs)
	for i := 0; i < numOfJobs; i++ {
		base := rand.Float32() * 100
		multiplier := rand.Float32() * 10
		product := Product{Base: base, Multiplier: multiplier}

		jobs[i] = &workerpool.Job{
			Descriptor: workerpool.JobDescriptor{
				ID:       workerpool.JobID(fmt.Sprintf("%v", i)),
				JType:    "anyType",
				Metadata: nil,
			},
			ExecFn: computePrice,
			Args:   product,
		}
	}
	return jobs
}
