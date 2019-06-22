package check

import (
	"context"
	"errors"
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

		span.AddAttributes(trace.BoolAttribute("done", op.Done))

		if op.Error != nil {
			span.Annotate([]trace.Attribute{
				trace.Int64Attribute("error.code", op.Error.Code),
				trace.StringAttribute("error.message", op.Error.Message),
				trace.StringAttribute("error.details", fmt.Sprint(op.Error.Details)),
			}, "Error")

			j, _ := op.Error.MarshalJSON()
			return false, errors.New(string(j))
		}

		return op.Done, nil
	}
}
