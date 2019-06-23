package deployments

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/internal/check"
	"go.opencensus.io/trace"
	"google.golang.org/api/deploymentmanager/v2"
	"k8s.io/apimachinery/pkg/util/wait"
)

func Ref(name, property string) string {
	return fmt.Sprintf("$(ref.%s.%s)", name, property)
}

func SelfLink(name string) string {
	return Ref(name, "selfLink")
}

type Metadata struct {
	DependsOn []string `json:"dependsOn,omitempty"`
}

type Output struct {
	Name  string `json:"name"`
	Value string `json:"name"`
}

type Resource struct {
	Name       string      `json:"name"`
	Type       string      `json:"type"`
	Properties interface{} `json:"properties"`
	Metadata   *Metadata   `json:"metadata,omitempty"`
	Outputs    []Output    `json:"outputs,omitempty"`
}

type Config struct {
	Resources []Resource `json:"resources,omitempty"`
}

func Insert(ctx context.Context, project, name string, config *Config, labels map[string]string, dryRun bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.internal.deployments.Insert")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("name", name),
		trace.BoolAttribute("dry_run", dryRun),
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

	if dryRun {
		b, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	}

	j, err := json.Marshal(config)
	if err != nil {
		return err
	}

	op, err := dm.Deployments.Insert(project, &deploymentmanager.Deployment{
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

	return wait.Poll(10*time.Second, 5*time.Minute, check.DM(ctx, dm, project, op.Name))
}

func Update(ctx context.Context, project, name string, config *Config, dryRun bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.internal.deployments.Update")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("name", name),
		trace.BoolAttribute("dry_run", dryRun),
	)
	defer span.End()

	dm, err := deploymentmanager.NewService(ctx)
	if err != nil {
		return err
	}

	if dryRun {
		b, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	}

	j, err := json.Marshal(config)
	if err != nil {
		return err
	}

	op, err := dm.Deployments.Patch(project, name, &deploymentmanager.Deployment{
		Target: &deploymentmanager.TargetConfiguration{
			Config: &deploymentmanager.ConfigFile{
				Content: string(j),
			},
		},
	}).Do()
	if err != nil {
		return err
	}

	return wait.Poll(10*time.Second, 5*time.Minute, check.DM(ctx, dm, project, op.Name))
}
func Delete(ctx context.Context, project, name string, dryRun bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.internal.deployments.Delete")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("name", name),
		trace.BoolAttribute("dry_run", dryRun),
	)
	defer span.End()

	dm, err := deploymentmanager.NewService(ctx)
	if err != nil {
		return err
	}

	if dryRun {
		return nil
	}

	op, err := dm.Deployments.Delete(project, name).Do()
	if err != nil {
		return err
	}

	return wait.Poll(10*time.Second, 5*time.Minute, check.DM(ctx, dm, project, op.Name))
}
