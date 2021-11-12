package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sync/atomic"

	"golang.org/x/sync/errgroup"
)

func main() {

	var choice int

	for {
		fmt.Println("Which context to use?")
		fmt.Print("Enter 1=no timetout 2=no timeout with error 3=with timeout 4=exit:")

		n, err := fmt.Scanf("%d", &choice)
		if n != 1 || err != nil {
			fmt.Println("Follow directions!")
			return
		}

		switch choice {
		case 1:

		case 2:

		case 3:

		case 4:
			os.Exit(1)
		}

	}
}

type User struct {
	Id   int64
	Name string
}

type Profile struct {
	Id     int64
	Userid int64
}

func GetFriendIds(user int64) {

}

func GetFriends(ctx context.Context, user int64) (map[string]*User, error) {
	// Produce
	var friendIds []int64
	for it := GetFriendIds(user); ; {
		if id, err := it.Next(ctx); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("GetFriendIds %d: %w", user, err)
		} else {
			friendIds = append(friendIds, id)
		}
	}

	// Map
	ret := map[string]*User{}
	for _, friendId := range friendIds {
		if friend, err := GetUserProfile(ctx, friendId); err != nil {
			return nil, fmt.Errorf("GetUserProfile %d: %w", friendId, err)
		} else {
			ret[friend.Name] = friend
		}
	}
	return ret, nil
}

func GetFriends2(ctx context.Context, user int64) (map[string]*User, error) {
	friendIds := make(chan int64)

	// Produce
	go func() {
		defer close(friendIds)
		for it := GetFriendIds(user); ; {
			if id, err := it.Next(ctx); err != nil {
				if err == io.EOF {
					break
				}
				// What to do here?
				log.Fatalf("GetFriendIds %d: %s", user, err)
			} else {
				friendIds <- id
			}
		}
	}()

	friends := make(chan *User)

	// Map
	workers := int32(nWorkers)
	for i := 0; i < nWorkers; i++ {
		go func() {
			defer func() {
				// Last one out closes shop
				if atomic.AddInt32(&workers, -1) == 0 {
					close(friends)
				}
			}()

			for id := range friendIds {
				if friend, err := GetUserProfile(ctx, id); err != nil {
					// What to do here?
					log.Fatalf("GetUserProfile %d: %s", user, err)
				} else {
					friends <- friend
				}
			}
		}()
	}

	// Reduce
	ret := map[string]*User{}
	for friend := range friends {
		ret[friend.Name] = friend
	}

	return ret, nil
}

func GetFriends3(ctx context.Context, user int64) (map[string]*User, error) {
	g, ctx := errgroup.WithContext(ctx)
	friendIds := make(chan int64)

	// Produce
	g.Go(func() error {
		defer close(friendIds)
		for it := GetFriendIds(user); ; {
			if id, err := it.Next(ctx); err != nil {
				if err == io.EOF {
					return nil
				}
				return fmt.Errorf("GetFriendIds %d: %s", user, err)
			} else {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case friendIds <- id:
				}
			}
		}
	})

	friends := make(chan *User)

	// Map
	workers := int32(nWorkers)
	for i := 0; i < nWorkers; i++ {
		g.Go(func() error {
			defer func() {
				// Last one out closes shop
				if atomic.AddInt32(&workers, -1) == 0 {
					close(friends)
				}
			}()

			for id := range friendIds {
				if friend, err := GetUserProfile(ctx, id); err != nil {
					return fmt.Errorf("GetUserProfile %d: %s", user, err)
				} else {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case friends <- friend:
					}
				}
			}
			return nil
		})
	}

	// Reduce
	ret := map[string]*User{}
	g.Go(func() error {
		for friend := range friends {
			ret[friend.Name] = friend
		}
		return nil
	})

	return ret, g.Wait()
}
