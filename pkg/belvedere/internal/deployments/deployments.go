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

// Ref returns a reference to the named resource's property.
// https://cloud.google.com/deployment-manager/docs/configuration/use-references
func Ref(name, property string) string {
	return fmt.Sprintf("$(ref.%s.%s)", name, property)
}

// SelfLink returns a reference to the named resource's SelfLink property.
func SelfLink(name string) string {
	return Ref(name, "selfLink")
}

// Resource represents a Deployment Manager resource.
type Resource struct {
	Name       string      `json:"name"`
	Type       string      `json:"type"`
	Properties interface{} `json:"properties"`
}

// Creates a new deployment with the given name, resources, and labels.
func Create(ctx context.Context, project, name string, resources []Resource, labels map[string]string, dryRun bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.internal.deployments.Create")
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

	d := map[string][]Resource{"resources": resources}

	if dryRun {
		b, err := json.MarshalIndent(d, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	}

	j, err := json.Marshal(d)
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

// Updates the given deployment to add, remove, or modify resources.
func Update(ctx context.Context, project, name string, resources []Resource, dryRun bool) error {
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

	d := map[string][]Resource{"resources": resources}

	if dryRun {
		b, err := json.MarshalIndent(d, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	}

	j, err := json.Marshal(d)
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

// Deletes the given deployment.
func Delete(ctx context.Context, project, name string, dryRun, async bool) error {
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

	if async {
		return nil
	}

	return wait.Poll(10*time.Second, 5*time.Minute, check.DM(ctx, dm, project, op.Name))
}

func Labels(labels []*deploymentmanager.DeploymentLabelEntry) map[string]string {
	m := make(map[string]string)
	for _, e := range labels {
		m[e.Key] = e.Value
	}
	return m
}
