package belvedere

import (
	"context"
	"fmt"

	"go.opencensus.io/trace"
	"google.golang.org/api/serviceusage/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func checkServiceUsageOperation(ctx context.Context, su *serviceusage.Service, op *serviceusage.Operation) wait.ConditionFunc {
	return func() (bool, error) {
		_, span := trace.StartSpan(ctx, "belvedere.CheckServiceUsageOperation")
		defer span.End()

		o, err := su.Operations.Get(op.Name).Do()
		if err != nil {
			return false, err
		}

		if o.Error != nil {
			span.Annotate([]trace.Attribute{
				trace.Int64Attribute("error.code", o.Error.Code),
				trace.StringAttribute("error.message", o.Error.Message),
				trace.StringAttribute("error.details", fmt.Sprint(o.Error.Details)),
			}, "Error during operation")
		}
		span.AddAttributes(trace.BoolAttribute("done", op.Done))
		return o.Done, nil
	}
}
