package check

import (
	"context"
	"fmt"

	"github.com/codahale/belvedere/pkg/belvedere/internal/waiter"
	"go.opencensus.io/trace"
	"google.golang.org/api/serviceusage/v1"
)

// SU returns a waiter.Condition for the given Service Usage operation completing.
func SU(ctx context.Context, su *serviceusage.Service, operation string) waiter.Condition {
	return func() (bool, error) {
		ctx, span := trace.StartSpan(ctx, "belvedere.internal.check.SU")
		defer span.End()

		span.AddAttributes(trace.StringAttribute("operation", operation))

		// Fetch the operation's status and any errors.
		op, err := su.Operations.Get(operation).Context(ctx).Do()
		if err != nil {
			return false, fmt.Errorf("error getting operation: %w", err)
		}

		span.AddAttributes(trace.BoolAttribute("done", op.Done))

		// Check for errors in the operation.
		if op.Error != nil {
			err := &failedOperationError{Message: op.Error}

			span.SetStatus(trace.Status{
				Code:    trace.StatusCodeInternal,
				Message: err.Error(),
			})

			return false, err
		}

		// Keep waiting unless the operation is done.
		return op.Done, nil
	}
}
