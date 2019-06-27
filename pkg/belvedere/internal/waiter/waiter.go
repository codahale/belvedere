package waiter

import (
	"context"
	"fmt"
	"time"
)

// Condition is a function which monitors an ongoing process and reports if the process is done or
// if there was an error monitoring the process or during the process itself.
type Condition func() (done bool, err error)

// WithInterval returns a new context with the given polling interval. This is required before using
// the Poll function.
func WithInterval(ctx context.Context, interval time.Duration) context.Context {
	return context.WithValue(ctx, intervalKey{}, interval)
}

// Poll checks the given condition using the given context's interval (see WithInterval) and
// deadline. If the condition completes, returns nil. If the condition returns an error, returns
// that error. If the context's deadline expires or the context is cancelled, returns the related
// error.
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
			return ctx.Err()
		}
	}
}

type intervalKey struct{}
