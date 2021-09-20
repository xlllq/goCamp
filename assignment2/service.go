package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"
)

func service_rpc1(ctx context.Context) error {
	//calling rpc...
	time.Sleep(4 * time.Second)

	//if called shutdown, then canceled
	select {
	case <-ctx.Done():
		fmt.Println("rpc1 canceled")
	default:
		fmt.Println("rpc1 Completed!")
	}
	//simulating rpc fail
	return errors.New("rpc1 failed")
}

func service_rpc2(ctx context.Context) error {
	userAgent, ok := ctx.Value(ctxKey("UserAgent")).(string)
	if ok {
		fmt.Println("UserAgent is", userAgent)
	} else {
		fmt.Println("failed to get UserAgent")
	}

	//calling rpc...
	time.Sleep(8 * time.Second)
	//

	//if another rpc failed, exit this goroutine
	select {
	case <-ctx.Done():
		fmt.Println("rpc2 canceled")
		return ctx.Err()
	default:
		//Should not arrive
		fmt.Println("rpc2 Completed!")
		return nil
	}
}

func service(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return service_rpc1(ctx)
	})
	group.Go(func() error {
		return service_rpc2(ctx)
	})

	return group.Wait()
}
