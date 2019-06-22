package check

import (
	"context"
	"errors"

	"go.opencensus.io/trace"
	"google.golang.org/api/compute/v0.beta"
	"k8s.io/apimachinery/pkg/util/wait"
)

func GCE(ctx context.Context, gce *compute.Service, project, operation string) wait.ConditionFunc {
	return func() (bool, error) {
		ctx, span := trace.StartSpan(ctx, "belvedere.internal.check.GCE")
		span.AddAttributes(trace.StringAttribute("project", project))
		span.AddAttributes(trace.StringAttribute("operation", operation))
		defer span.End()

		op, err := gce.GlobalOperations.Get(project, operation).Context(ctx).Do()
		if err != nil {
			return false, err
		}
		span.AddAttributes(trace.StringAttribute("status", op.Status))

		if op.Error != nil {
			for _, e := range op.Error.Errors {
				span.Annotate([]trace.Attribute{
					trace.StringAttribute("code", e.Code),
					trace.StringAttribute("message", e.Message),
					trace.StringAttribute("location", e.Location),
				}, "Error")
			}
			j, _ := op.Error.MarshalJSON()
			return false, errors.New(string(j))
		}

		return op.Status == "DONE", nil
	}
}
