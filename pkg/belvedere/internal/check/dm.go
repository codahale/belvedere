package check

import (
	"context"

	"go.opencensus.io/trace"
	"google.golang.org/api/deploymentmanager/v2"
	"k8s.io/apimachinery/pkg/util/wait"
)

func DM(ctx context.Context, dm *deploymentmanager.Service, project string, operation string) wait.ConditionFunc {
	return func() (bool, error) {
		ctx, span := trace.StartSpan(ctx, "belvedere.internal.check.DM")
		span.AddAttributes(trace.StringAttribute("project", project))
		span.AddAttributes(trace.StringAttribute("operation", operation))
		defer span.End()

		op, err := dm.Operations.Get(project, operation).Context(ctx).Do()
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
