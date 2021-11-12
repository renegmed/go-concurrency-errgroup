package pipeline

import (
	"context"
	"fmt"
	"learn-concurrency/errgroup/catching-panic-scoffman/errgroup"
	//"golang.org/x/sync/errgroup"
)

type Executor func(interface{}) (interface{}, error)

type Pipeline interface {
	Pipe(executor Executor) Pipeline
	Merge() <-chan interface{}
}

type pipeline struct {
	dataC     chan interface{}
	errC      chan error
	executors []Executor
	ctx       context.Context
}

func New(ctx context.Context, f func(chan interface{})) Pipeline {
	inC := make(chan interface{})

	go f(inC)

	return &pipeline{
		dataC:     inC,
		errC:      make(chan error),
		executors: []Executor{},
		ctx:       ctx,
	}
}

func (p *pipeline) Pipe(executor Executor) Pipeline {
	p.executors = append(p.executors, executor)

	return p
}

func (p *pipeline) Merge() <-chan interface{} {
	outC := make(chan interface{})
	errC := make(chan error)

	g, ctx := errgroup.WithContext(p.ctx)

	for i := 0; i < len(p.executors); i++ {
		//p.dataC, p.errC = run(ctx, p.dataC, p.executors[i])

		g.Go(func() error {
			defer close(p.dataC)
			for {
				select {
				case <-ctx.Done():
					fmt.Printf("... ctx.Done %v\n", ctx.Err())
					return ctx.Err()
				case v, ok := <-p.dataC:
					if !ok {
						fmt.Println("...no more function")
						return nil
					}

					res, err := p.executors[i](v)
					if err != nil {
						fmt.Println("... error on executor: ", err)
						errC <- err
					}
					outC <- res
				}
			}
			// for v := range p.dataC {
			// 	res, err := p.executors[i](v)
			// 	if err != nil {
			// 		errC <- err
			// 		continue
			// 	}

			// 	outC <- res
			// }
		})
	}

	go func() {
		err := g.Wait()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		close(outC)
		close(errC)

	}()
	return p.dataC
}

// func run(ctx context.Context,
// 	inC <-chan interface{},
// 	f Executor) (chan interface{}, chan error) {
// 	outC := make(chan interface{})
// 	errC := make(chan error)

// 	g.Go(func() error {
// 		defer close(outC)
// 		for v := range inC {
// 			res, err := f(v)
// 			if err != nil {
// 				errC <- err
// 				continue
// 			}

// 			outC <- res
// 		}
// 	})

// 	return outC, errC
// }
