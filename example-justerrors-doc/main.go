/*

JustErrors illustrates the use of a Group in place of a sync.WaitGroup
to simplify goroutine counting and error handling.

This example is derived from the sync.WaitGroup example
at https://golang.org/pkg/sync/#example_WaitGroup.

*/

package main

import (
	"fmt"
	"net/http"
	"sync"

	"golang.org/x/sync/errgroup"
)

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
		withErrGroup()
	case 2:
		withoutErrGroup()
	}

}

func withErrGroup() {
	g := new(errgroup.Group)
	var urls = []string{
		"http://www.golang.org/",
		"http://www.google.com/",
		"http://www.somestupidname.com/",
	}
	for _, url := range urls {
		// Launch a goroutine to fetch the URL.
		url := url // https://golang.org/doc/faq#closures_and_goroutines, there would be a data race between multiple goroutines, because the for loop reuses url in each iteration
		g.Go(func() error {
			// Fetch the URL.
			resp, err := http.Get(url)
			if err == nil {
				fmt.Println("...url:", url)
				resp.Body.Close()
				return nil
			}
			fmt.Println("...url with error:", url)
			return err
		})
	}
	// Wait for all HTTP fetches to complete.
	// if err := g.Wait(); err == nil {
	// 	fmt.Println("Successfully fetched all URLs.")
	// }
	if err := g.Wait(); err != nil {
		fmt.Println("...error:", err)
		return
	}
	fmt.Println("Successfully fetched all URLs.")

}

type httpPkg struct{}

func (httpPkg) Get(url string) {
	resp, err := http.Get(url)
	if err == nil {
		fmt.Println("...url:", url)
		resp.Body.Close()
		return
	}
	fmt.Printf("...url %swith error: %v\n", url, err)
}

func withoutErrGroup() {

	var h httpPkg
	var wg sync.WaitGroup
	var urls = []string{
		"http://www.golang.org/",
		"http://www.google.com/",
		"http://www.somestupidname.com/",
	}
	for _, url := range urls {
		// Increment the WaitGroup counter.
		wg.Add(1)
		// Launch a goroutine to fetch the URL.
		go func(url string) {
			// Decrement the counter when the goroutine completes.
			defer wg.Done()
			// Fetch the URL.
			h.Get(url)

		}(url)
	}
	// Wait for all HTTP fetches to complete.
	wg.Wait()
	fmt.Println("Successfully fetched all URLs.")
}

/*

// https://pkg.go.dev/sync#example-WaitGroup

package main

import (
	"sync"
)

type httpPkg struct{}

func (httpPkg) Get(url string) {}

var http httpPkg

func main() {
	var wg sync.WaitGroup
	var urls = []string{
		"http://www.golang.org/",
		"http://www.google.com/",
		"http://www.somestupidname.com/",
	}
	for _, url := range urls {
		// Increment the WaitGroup counter.
		wg.Add(1)
		// Launch a goroutine to fetch the URL.
		go func(url string) {
			// Decrement the counter when the goroutine completes.
			defer wg.Done()
			// Fetch the URL.
			http.Get(url)
		}(url)
	}
	// Wait for all HTTP fetches to complete.
	wg.Wait()
}
*/
