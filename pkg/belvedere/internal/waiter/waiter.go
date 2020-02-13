package waiter

import (
	"context"
	"time"
)

// Condition is a function which monitors an ongoing process and reports if the process is done or
// if there was an error monitoring the process or during the process itself.
type Condition func() (done bool, err error)

// Poll checks the given condition using the given interval and deadline. If the condition
// completes, returns nil. If the condition returns an error, returns that error. If the context's
// deadline expires or the context is cancelled, returns the related error.
func Poll(ctx context.Context, interval time.Duration, c Condition) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			done, err := c()
			if err != nil {
				return err
			}

			if done {
				return nil
			}
		}
	}
}
