package deployments

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.opencensus.io/trace"
	"google.golang.org/api/deploymentmanager/v2"
	"k8s.io/apimachinery/pkg/util/wait"
)

func Create(ctx context.Context, projectID, name string, config *Config, labels map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.createDeployment")
	span.AddAttributes(
		trace.StringAttribute("project_id", projectID),
		trace.StringAttribute("name", name),
	)
	defer span.End()

	dm, err := deploymentmanager.NewService(ctx)
	if err != nil {
		return err
	}

	var l []*deploymentmanager.DeploymentLabelEntry
	for k, v := range labels {
		l = append(l, &deploymentmanager.DeploymentLabelEntry{
			Key:   k,
			Value: v,
		})
	}

	j, err := json.Marshal(config)
	if err != nil {
		return err
	}

	op, err := dm.Deployments.Insert(projectID, &deploymentmanager.Deployment{
		Labels: l,
		Name:   name,
		Target: &deploymentmanager.TargetConfiguration{
			Config: &deploymentmanager.ConfigFile{
				Content: string(j),
			},
		},
	}).Do()
	if err != nil {
		return err
	}

	f := checkOperation(ctx, dm, projectID, op.Name)
	return wait.Poll(10*time.Second, 5*time.Minute, f)
}

func checkOperation(ctx context.Context, dm *deploymentmanager.Service, projectID string, operation string) wait.ConditionFunc {
	return func() (bool, error) {
		ctx, span := trace.StartSpan(ctx, "belvedere.internal.deployments.checkOperation")
		span.AddAttributes(trace.StringAttribute("operation", operation))
		defer span.End()

		op, err := dm.Operations.Get(projectID, operation).Context(ctx).Do()
		if err != nil {
			return false, err
		}

		if op.Error != nil {
			for i, e := range op.Error.Errors {
				prefix := fmt.Sprintf("error.%d.", i)
				span.Annotate([]trace.Attribute{
					trace.StringAttribute(prefix+"code", e.Code),
					trace.StringAttribute(prefix+"message", e.Message),
					trace.StringAttribute(prefix+"location", e.Location),
				}, "Error")
			}
			span.SetStatus(trace.Status{Code: trace.StatusCodeAborted})
		}
		span.AddAttributes(trace.StringAttribute("status", op.Status))
		return op.Status == "DONE", nil
	}
}
