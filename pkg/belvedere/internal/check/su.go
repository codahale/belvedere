package check

import (
	"context"
	"errors"
	"fmt"

	"github.com/codahale/belvedere/pkg/belvedere/internal/waiter"
	"go.opencensus.io/trace"
	"google.golang.org/api/serviceusage/v1"
)

// GCE returns a handle for a Service Usage operation.
func SU(ctx context.Context, su *serviceusage.Service, operation string) waiter.Condition {
	return func() (bool, error) {
		ctx, span := trace.StartSpan(ctx, "belvedere.internal.check.SU")
		span.AddAttributes(trace.StringAttribute("operation", operation))
		defer span.End()

		// Fetch the operation's status and any errors.
		op, err := su.Operations.Get(operation).Context(ctx).
			Fields("status", "error").Do()
		if err != nil {
			return false, err
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
			return false, errors.New(string(j))
		}

		// Keep waiting unless the operation is done.
		return op.Done, nil
	}
}
