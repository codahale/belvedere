package waiter

import (
	"context"
	"fmt"
	"time"
)

type Condition func() (done bool, err error)

type intervalKey struct{}

func WithInterval(ctx context.Context, interval time.Duration) context.Context {
	return context.WithValue(ctx, intervalKey{}, interval)
}

func Poll(ctx context.Context, c Condition) error {
	interval, ok := ctx.Value(intervalKey{}).(time.Duration)
	if !ok {
		return fmt.Errorf("no interval set")
	}

	for {
		select {
		case <-time.After(interval):
			done, err := c()
			if err != nil {
				return err
			}

			if done {
				return nil
			}
		case <-ctx.Done():
			return fmt.Errorf("wait cancelled")
		}
	}
}
