package check

import (
	"context"
	"fmt"

	"github.com/codahale/belvedere/pkg/belvedere/internal/gcp"
	"github.com/codahale/belvedere/pkg/belvedere/internal/waiter"
	"go.opencensus.io/trace"
)

// SU returns a waiter.Condition for the given Service Usage operation completing.
func SU(ctx context.Context, operation string) waiter.Condition {
	return func() (bool, error) {
		ctx, span := trace.StartSpan(ctx, "belvedere.internal.check.SU")
		span.AddAttributes(trace.StringAttribute("operation", operation))
		defer span.End()

		// Get or create our SU client.
		su, err := gcp.ServiceUsage(ctx)
		if err != nil {
			return false, err
		}

		// Fetch the operation's status and any errors.
		op, err := su.Operations.Get(operation).Context(ctx).Do()
		if err != nil {
			return false, fmt.Errorf("error getting operation: %w", err)
		}
		span.AddAttributes(trace.BoolAttribute("done", op.Done))

		// Check for errors in the operation.
		if op.Error != nil {
			// Record the error as an annotation.
			span.Annotate([]trace.Attribute{
				trace.Int64Attribute("error.code", op.Error.Code),
				trace.StringAttribute("error.message", op.Error.Message),
				trace.StringAttribute("error.details", fmt.Sprint(op.Error.Details)),
			}, "Error")

			// Exit with a maximally descriptive error.
			j, _ := op.Error.MarshalJSON()
			return false, fmt.Errorf("operation failed: %s", j)
		}

		// Keep waiting unless the operation is done.
		return op.Done, nil
	}
}
