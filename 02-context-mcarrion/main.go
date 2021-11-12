package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/sync/errgroup"
)

func main() {
	var files = []string{"file1.csv", "file2.csv", "file3.csv"}
	var files2 = []string{"file1.csv", "file2.csv", "file3.csv", "file4.csv"}

	var ctx context.Context
	var cancel context.CancelFunc

	var choice int
	//var wait = make(chan struct{}, 1)

	for {
		fmt.Println("Which context to use?")
		fmt.Print("Enter 1=no timetout 2=no timeout with error 3=with timeout 4=exit:")

		n, err := fmt.Scanf("%d", &choice)
		if n != 1 || err != nil {
			fmt.Println("Follow directions!")
			return
		}

		var f []string

		switch choice {
		case 1:
			ctx = context.Background()
			f = files
		case 2:
			ctx = context.Background()
			f = files2
		case 3:
			ctx, cancel = context.WithTimeout(context.Background(), 1*time.Millisecond)
			f = files
			defer cancel()
		case 4:
			os.Exit(1)
		}

		// wait := errGroup(ctx)
		wait := errGroup(ctx, f)

		<-wait

		fmt.Println("----------------")
	}

}

func errGroup(ctx context.Context, files []string) chan struct{} {
	ch := make(chan struct{}, 1)

	g, ctx := errgroup.WithContext(ctx)

	// for _, file := range []string{"file1.csv", "file2.csv", "file3.csv", "file4.csv"} {
	for _, file := range files {
		file := file // copy for local scope

		g.Go(func() error { // read and process one file
			ch2, err := read(file)
			if err != nil {
				return fmt.Errorf("error reading %w", err)
			}

			for {
				select {
				case <-ctx.Done():
					fmt.Printf("Context completed %v\n", ctx.Err())

					return ctx.Err()
				case line, ok := <-ch2:
					if !ok {
						return nil
					}

					fmt.Println(line)
				}
			}
		})
	}

	go func() {
		if err := g.Wait(); err != nil { // wait until all processing is done, print all errors
			fmt.Printf("Error reading files: %v\n", err)
		}
		close(ch)
	}()

	return ch
}

func read(file string) (<-chan []string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("opening file %w", err)
	}

	ch := make(chan []string)

	go func() {
		cr := csv.NewReader(f)

		time.Sleep(time.Millisecond) // XXX: Intentional sleep 😴

		for {
			record, err := cr.Read()
			if errors.Is(err, io.EOF) {
				close(ch)

				return
			}

			ch <- record
		}
	}()

	return ch, nil
}
