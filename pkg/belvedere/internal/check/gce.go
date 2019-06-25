package check

import (
	"context"
	"errors"

	"github.com/codahale/belvedere/pkg/belvedere/internal/gcp"
	"github.com/codahale/belvedere/pkg/belvedere/internal/waiter"
	"go.opencensus.io/trace"
)

// GCE returns a handle for a global GCE operation.
func GCE(ctx context.Context, project, operation string) waiter.Condition {
	return func() (bool, error) {
		ctx, span := trace.StartSpan(ctx, "belvedere.internal.check.GCE")
		span.AddAttributes(trace.StringAttribute("project", project))
		span.AddAttributes(trace.StringAttribute("operation", operation))
		defer span.End()

		ctx, gce, err := gcp.Compute(ctx)
		if err != nil {
			return false, err
		}

		// Fetch the operation's status and any errors.
		op, err := gce.GlobalOperations.Get(project, operation).Context(ctx).
			Fields("status", "error").Do()
		if err != nil {
			return false, err
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
			return false, errors.New(string(j))
		}

		// Keep waiting unless the operation is done.
		return op.Status == "DONE", nil
	}
}
