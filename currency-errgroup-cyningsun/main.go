package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"golang.org/x/sync/errgroup"
)

var (
	Blogger = find("1227368500")
	Weibo   = find("H3GIgngon")
)

type Result string

type Find func(ctx context.Context, query string) (Result, error)

func find(kind string) Find {
	return func(_ context.Context, query string) (Result, error) {

		i := rand.Intn(5)
		if i > 1 {
			return Result(fmt.Sprintf("%s result for %q", kind, query)), nil
		}

		return "", fmt.Errorf("...error from find: %s, %d", kind, i)
	}
}

func main() {

	rand.Seed(time.Now().UnixNano())

	SinaWeibo := func(ctx context.Context, query string) ([]Result, error) {

		g, ctx := errgroup.WithContext(ctx)

		finds := []Find{Blogger, Weibo}
		results := make([]Result, len(finds))

		for i, find := range finds {
			i, find := i, find // https://golang.org/doc/faq#closures_and_goroutines, go vet

			g.Go(func() error {

				result, err := find(ctx, query)
				if err == nil {
					results[i] = result
				}
				return err
			})
		}
		if err := g.Wait(); err != nil {
			fmt.Println("...error:", err)
			return nil, err
		}
		return results, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)

	results, err := SinaWeibo(ctx, "https://weibo.com/1227368500/H3GIgngon")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	for _, result := range results {
		fmt.Println(result)
	}
	cancel()
}
