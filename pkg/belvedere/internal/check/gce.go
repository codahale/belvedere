package check

import (
	"context"

	"go.opencensus.io/trace"
	compute "google.golang.org/api/compute/v0.beta"
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

		if op.Error != nil {
			for _, e := range op.Error.Errors {
				span.Annotate([]trace.Attribute{
					trace.StringAttribute("code", e.Code),
					trace.StringAttribute("message", e.Message),
					trace.StringAttribute("location", e.Location),
				}, "Error")
			}
			span.SetStatus(trace.Status{Code: trace.StatusCodeAborted})
		}
		span.AddAttributes(trace.StringAttribute("status", op.Status))
		return op.Status == "DONE", nil
	}
}
