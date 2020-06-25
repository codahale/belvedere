//nolint:dupl // duplicated code b/c no type parameters
package check

import (
	"context"
	"fmt"

	"github.com/codahale/belvedere/pkg/belvedere/internal/waiter"
	"go.opencensus.io/trace"
	compute "google.golang.org/api/compute/v1"
)

// GCE returns a waiter.Condition for the given Compute Engine operation completing.
func GCE(ctx context.Context, gce *compute.Service, project, operation string) waiter.Condition {
	return func() (bool, error) {
		ctx, span := trace.StartSpan(ctx, "belvedere.internal.check.GCE")
		defer span.End()

		span.AddAttributes(
			trace.StringAttribute("project", project),
			trace.StringAttribute("operation", operation),
		)

		// Fetch the operation's status and any errors.
		op, err := gce.GlobalOperations.Get(project, operation).Context(ctx).
			Fields("status", "error").Do()
		if err != nil {
			return false, fmt.Errorf("error getting operation: %w", err)
		}

		span.AddAttributes(trace.StringAttribute("status", op.Status))

		// Check for errors in the operation.
		if op.Error != nil {
			// Record all errors as annotations.
			for _, e := range op.Error.Errors {
				span.Annotate([]trace.Attribute{
					trace.StringAttribute("code", e.Code),
					trace.StringAttribute("message", e.Message),
					trace.StringAttribute("location", e.Location),
				}, "Error")
			}

			// Exit with a maximally descriptive error.
			return false, &FailedOperationError{Message: op.Error}
		}

		// Keep waiting unless the operation is done.
		return op.Status == "DONE", nil
	}
}
