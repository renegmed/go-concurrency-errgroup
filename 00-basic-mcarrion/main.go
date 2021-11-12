package main

import (
	"context"
	"errors"
	"log"

	"golang.org/x/sync/errgroup"
)

func main() {
	messages := make(chan string)

	g, ctx := errgroup.WithContext(context.Background())

	g.Go(func() error {
		defer close(messages)

		for _, msg := range []string{"a", "b", "c", "d"} {
			select {
			case messages <- msg:
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		return nil
	})

	g.Go(func() error {
		for {
			select {
			case msg, open := <-messages:
				if !open {
					return nil
				}

				if msg == "c" {
					return errors.New("I don't like c")
				}

			case <-ctx.Done():
				return ctx.Err()
			}
		}

		//return nil
	})

	if err := g.Wait(); err != nil {
		log.Fatalln("wait", err)
	}
}
