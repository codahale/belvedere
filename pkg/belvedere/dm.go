package belvedere

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.opencensus.io/trace"
	"google.golang.org/api/deploymentmanager/v2"
	"k8s.io/apimachinery/pkg/util/wait"
)

func ref(name, property string) string {
	return fmt.Sprintf("$(ref.%s.%s)", name, property)
}

func selfLink(name string) string {
	return ref(name, "selfLink")
}

type metadata struct {
	DependsOn []string `json:"dependsOn,omitempty"`
}

type output struct {
	Name  string `json:"name"`
	Value string `json:"name"`
}

type resource struct {
	Name       string      `json:"name"`
	Type       string      `json:"type"`
	Properties interface{} `json:"properties"`
	Metadata   *metadata   `json:"metadata,omitempty"`
	Outputs    []output    `json:"outputs,omitempty"`
}

type config struct {
	Resources []resource `json:"resources,omitempty"`
}

type deployment struct {
	projectID string
	name      string
	labels    map[string]string
	config    config
}

func createDeployment(ctx context.Context, deployment deployment) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.createDeployment")
	span.AddAttributes(
		trace.StringAttribute("project_id", deployment.projectID),
		trace.StringAttribute("name", deployment.name),
	)
	defer span.End()

	dm, err := deploymentmanager.NewService(ctx)
	if err != nil {
		return err
	}

	var labels []*deploymentmanager.DeploymentLabelEntry
	for k, v := range deployment.labels {
		labels = append(labels, &deploymentmanager.DeploymentLabelEntry{
			Key:   k,
			Value: v,
		})
	}

	j, err := json.Marshal(deployment.config)
	if err != nil {
		return err
	}

	op, err := dm.Deployments.Insert(deployment.projectID, &deploymentmanager.Deployment{
		Labels: labels,
		Name:   deployment.name,
		Target: &deploymentmanager.TargetConfiguration{
			Config: &deploymentmanager.ConfigFile{
				Content: string(j),
			},
		},
	}).Do()
	if err != nil {
		return err
	}

	f := checkDMOperation(ctx, dm, deployment.projectID, op)
	return wait.Poll(10*time.Second, 5*time.Minute, f)
}

func deleteDeployment(ctx context.Context, projectID, name string) error {
	dm, err := deploymentmanager.NewService(ctx)
	if err != nil {
		return err
	}

	op, err := dm.Deployments.Delete(projectID, name).Do()
	if err != nil {
		return err
	}

	f := checkDMOperation(ctx, dm, projectID, op)
	return wait.Poll(10*time.Second, 5*time.Minute, f)
}
func checkDMOperation(ctx context.Context, dm *deploymentmanager.Service, projectID string, op *deploymentmanager.Operation) wait.ConditionFunc {
	return func() (bool, error) {
		_, span := trace.StartSpan(ctx, "belvedere.checkDMOperation")
		defer span.End()

		o, err := dm.Operations.Get(projectID, op.Name).Do()
		if err != nil {
			return false, err
		}

		if o.Error != nil {
			for i, e := range o.Error.Errors {
				prefix := fmt.Sprintf("error.%d.", i)
				span.Annotate([]trace.Attribute{
					trace.StringAttribute(prefix+"code", e.Code),
					trace.StringAttribute(prefix+"message", e.Message),
					trace.StringAttribute(prefix+"location", e.Location),
				}, "Error")
			}
			span.SetStatus(trace.Status{Code: trace.StatusCodeAborted})
		}
		span.AddAttributes(trace.StringAttribute("status", o.Status))
		return o.Status == "DONE", nil
	}
}
