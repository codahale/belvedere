package check

import (
	"context"
	"fmt"

	"go.opencensus.io/trace"
	"google.golang.org/api/serviceusage/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func SU(ctx context.Context, su *serviceusage.Service, operation string) wait.ConditionFunc {
	return func() (bool, error) {
		ctx, span := trace.StartSpan(ctx, "belvedere.internal.check.SU")
		span.AddAttributes(trace.StringAttribute("operation", operation))
		defer span.End()

		op, err := su.Operations.Get(operation).Context(ctx).Do()
		if err != nil {
			return false, err
		}

		if op.Error != nil {
			span.Annotate([]trace.Attribute{
				trace.Int64Attribute("error.code", op.Error.Code),
				trace.StringAttribute("error.message", op.Error.Message),
				trace.StringAttribute("error.details", fmt.Sprint(op.Error.Details)),
			}, "Error")
			span.SetStatus(trace.Status{Code: trace.StatusCodeAborted, Message: op.Error.Message})
		}
		span.AddAttributes(trace.BoolAttribute("done", op.Done))
		return op.Done, nil
	}
}
