package gcp

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
	"google.golang.org/api/googleapi"
)

// ModifyLoop performs the given operation in a retry loop with exponential backoff, retrying if the
// operation returns a 409 Conflict response from a GCP API. This is a required primitive for
// modifying IAM policies safely.
func ModifyLoop(interval, timeout time.Duration, f func() error) error {
	bo := backoff.NewExponentialBackOff()
	bo.MaxInterval = interval
	bo.MaxElapsedTime = timeout

	for {
		// Perform operation and check to see if the error is retryable.
		if err := f(); isRetryable(err) {
			d := bo.NextBackOff()
			if d == backoff.Stop {
				// If the total time has elapsed, return an error.
				return context.DeadlineExceeded
			}

			// Otherwise, wait for the backoff period and retry later.
			time.Sleep(d)

			continue
		} else if err != nil {
			// If the operation resulted in an error, exit.
			return err
		}

		// If the operation was successful, exit.
		return nil
	}
}

func isRetryable(err error) bool {
	// If the operation resulted in a conflict, back off and retry.
	var googleErr *googleapi.Error
	if errors.As(err, &googleErr) {
		if googleErr.Code == http.StatusConflict || googleErr.Code == http.StatusPreconditionFailed {
			return true
		}
	}

	return false
}
