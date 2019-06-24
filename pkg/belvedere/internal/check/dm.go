package check

import (
	"context"
	"errors"

	"go.opencensus.io/trace"
	"google.golang.org/api/deploymentmanager/v2"
	"k8s.io/apimachinery/pkg/util/wait"
)

// DM returns a ConditionFunc which checks the progress of a Deployment Manager operation.
func DM(ctx context.Context, dm *deploymentmanager.Service, project string, operation string) wait.ConditionFunc {
	return func() (bool, error) {
		ctx, span := trace.StartSpan(ctx, "belvedere.internal.check.DM")
		span.AddAttributes(trace.StringAttribute("project", project))
		span.AddAttributes(trace.StringAttribute("operation", operation))
		defer span.End()

		// Fetch the operation's status and any errors.
		op, err := dm.Operations.Get(project, operation).Context(ctx).
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
