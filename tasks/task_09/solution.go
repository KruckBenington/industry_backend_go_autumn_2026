package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type Result[R any] struct {
	Value R
	Err   error
}

var ErrInvalidWorkers = errors.New("workers must be positive")

func ParallelMap[T any, R any](ctx context.Context, workers int, in []T, fn func(context.Context, T) (R, error)) ([]Result[R], error) {
	if workers <= 0 {
		return nil, ErrInvalidWorkers
	}

	if ctx == nil || fn == nil {
		return nil, fmt.Errorf("doesn't supported nil values of ctx: %s", ctx)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if len(in) == 0 {
		return []Result[R]{}, nil
	}

	jobsChan := make(chan int)
	resultArray := make([]Result[R], len(in), len(in))
	var wg sync.WaitGroup

	for range min(workers, len(in)) {
		wg.Go(func() {
			for i := range jobsChan {
				result, err := fn(ctx, in[i])
				currentValue := result
				if err != nil {
					currentValue = *new(R)
				}
				resultArray[i] = Result[R]{Value: currentValue, Err: err}
			}
		},
		)
	}

leaveHere:
	for i := range in {
		if ctx.Err() != nil {
			break
		}

		select {
		case <-ctx.Done():
			break leaveHere
		case jobsChan <- i:
		}

	}
	close(jobsChan)
	wg.Wait()

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return resultArray, ctx.Err()
}
