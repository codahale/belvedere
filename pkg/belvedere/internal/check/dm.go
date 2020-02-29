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
		span.AddAttributes(
			trace.StringAttribute("project", project),
			trace.StringAttribute("operation", operation),
		)
		defer span.End()

		// Fetch the operation's status and any errors.
		op, err := dm.Operations.Get(project, operation).Context(ctx).
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
			j, _ := op.Error.MarshalJSON()
			return false, fmt.Errorf("operation failed: %s", j)
		}

		// Keep waiting unless the operation is done.
		return op.Status == "DONE", nil
	}
}
