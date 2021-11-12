/*

Parallel illustrates the use of a Group for synchronizing a simple parallel task: the
"Google Search 2.0" function from https://talks.golang.org/2012/concurrency.slide#46,
augmented with a Context and error-handling.

*/
package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	"golang.org/x/sync/errgroup"
)

var (
	Web   = fakeSearch("web")
	Image = fakeSearch("image")
	Video = fakeSearch("video")
)

type Result string
type Search func(ctx context.Context, query string) (Result, error)

func fakeSearch(kind string) Search {
	return func(_ context.Context, query string) (Result, error) {
		return Result(fmt.Sprintf("%s result for %q", kind, query)), nil
	}
}

func main() {
	var choice int
	fmt.Println("Which one to use?")
	fmt.Print("Enter 1=errGroup 2=no errGroup:")

	n, err := fmt.Scanf("%d", &choice)
	if n != 1 || err != nil {
		fmt.Println("Follow directions!")
		return
	}

	switch choice {
	case 1:
		useErrGroup()
	case 2:
		noErrGroup()
	}
}

func useErrGroup() {
	Google := func(ctx context.Context, query string) ([]Result, error) {
		g, ctx := errgroup.WithContext(ctx)

		searches := []Search{Web, Image, Video}

		results := make([]Result, len(searches)) // create array with specific size

		for i, search := range searches {
			i, search := i, search // https://golang.org/doc/faq#closures_and_goroutines, go vet
			g.Go(func() error {
				result, err := search(ctx, query)
				if err == nil {
					results[i] = result
				}
				return err
			})
		}
		if err := g.Wait(); err != nil {
			return nil, err
		}
		return results, nil
	}

	results, err := Google(context.Background(), "golang")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	for _, result := range results {
		fmt.Println(result)
	}
}

var (
	Web2   = fakeSearch2("web")
	Image2 = fakeSearch2("image")
	Video2 = fakeSearch2("video")
)

type Result2 string
type Search2 func(query string) (Result2, error)

func fakeSearch2(kind string) Search2 {
	return func(query string) (Result2, error) {
		return Result2(fmt.Sprintf("%s result for %q", kind, query)), nil
	}
}

func noErrGroup() {
	Google := func(query string) ([]Result2, error) {

		var wg sync.WaitGroup

		searches := []Search2{Web2, Image2, Video2}

		results := make([]Result2, len(searches)) // create array with specific size

		for i, search := range searches {
			i, search := i, search // https://golang.org/doc/faq#closures_and_goroutines, go vet
			wg.Add(1)
			go func() {
				defer wg.Done()
				result, err := search(query)
				if err == nil {
					results[i] = result
				}
			}()
		}
		wg.Wait()

		//fmt.Println(results)

		return results, nil
	}

	results, err := Google("golang")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	//fmt.Println("...results:\n", results)
	for _, result := range results {
		fmt.Println(result)
	}

}
