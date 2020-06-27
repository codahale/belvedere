package check

import (
	"context"
	"fmt"

	"github.com/codahale/belvedere/pkg/belvedere/internal/waiter"
	"go.opencensus.io/trace"
	"google.golang.org/api/deploymentmanager/v2"
)

// DM returns a waiter.Condition for the given Deployment Manager operation completing.
func DM(ctx context.Context, dm *deploymentmanager.Service, project string, operation string) waiter.Condition {
	return func() (bool, error) {
		ctx, span := trace.StartSpan(ctx, "belvedere.internal.check.DM")
		defer span.End()

		span.AddAttributes(
			trace.StringAttribute("project", project),
			trace.StringAttribute("operation", operation),
		)

		// Fetch the operation's status and any errors.
		op, err := dm.Operations.Get(project, operation).Context(ctx).
			Fields("status", "error").Do()
		if err != nil {
			return false, fmt.Errorf("error getting operation: %w", err)
		}

		span.AddAttributes(trace.StringAttribute("status", op.Status))

		// Check for errors in the operation.
		if op.Error != nil {
			return false, &failedOperationError{Message: op.Error}
		}

		// Keep waiting unless the operation is done.
		return op.Status == "DONE", nil
	}
}
