package gcp

import (
	"fmt"
	"time"

	"github.com/cenkalti/backoff"
	"google.golang.org/api/googleapi"
)

// ModifyLoop performs the given operation in a retry loop with exponential backoff, retrying if the
// operation returns a 409 Conflict response from a GCP API.
func ModifyLoop(interval, timeout time.Duration, f func() error) error {
	bo := backoff.NewExponentialBackOff()
	bo.MaxInterval = interval
	bo.MaxElapsedTime = timeout
	for {
		// Perform operation.
		err := f()

		if e, ok := err.(*googleapi.Error); ok {
			// If the operation resulted in a conflict, back off and retry.
			if e.Code == 409 {
				d := bo.NextBackOff()
				if d == backoff.Stop {
					// If the total time has elapsed, return an error.
					return fmt.Errorf("timeout after %s", bo.GetElapsedTime())
				}
				time.Sleep(d)
				continue
			}
		} else if err != nil {
			// If the operation resulted in an error, exit.
			return err
		}

		// If the operation was successful, exit.
		return nil
	}
}
