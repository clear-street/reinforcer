package main

import (
	"context"
	"fmt"
	"time"

	"github.com/clear-street/reinforcer/example/client"
	"github.com/clear-street/reinforcer/example/client/reinforced"
	"github.com/clear-street/reinforcer/pkg/runner"
	"github.com/slok/goresilience/retry"
	"github.com/slok/goresilience/timeout"
)

func main() {
	cl := client.NewClient()
	f := runner.NewFactory(
		timeout.NewMiddleware(timeout.Config{Timeout: 100 * time.Millisecond}),
		retry.NewMiddleware(retry.Config{
			Times: 10,
		}),
	)
	rCl := reinforced.NewClient(cl, f, reinforced.WithRetryableErrorPredicate(func(s string, err error) bool {
		// Always retry SayHello, don't retry any other error
		return s == reinforced.ClientMethods.SayHello
	}))
	for i := 0; i < 100; i++ {
		err := rCl.SayHello(context.Background(), "Christian")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
}
