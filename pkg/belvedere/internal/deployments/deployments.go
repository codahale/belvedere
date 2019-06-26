package deployments

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/codahale/belvedere/pkg/belvedere/internal/check"
	"github.com/codahale/belvedere/pkg/belvedere/internal/gcp"
	"github.com/codahale/belvedere/pkg/belvedere/internal/waiter"
	"go.opencensus.io/trace"
	"google.golang.org/api/deploymentmanager/v2"
	"google.golang.org/api/dns/v1"
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
	Name       string         `json:"name"`
	Type       string         `json:"type"`
	Properties json.Marshaler `json:"properties"`
}

type ServiceAccount struct {
	AccountID   string `json:"accountId"`
	DisplayName string `json:"displayName"`
}

func (s *ServiceAccount) MarshalJSON() ([]byte, error) {
	type NoMethod ServiceAccount
	raw := NoMethod(*s)
	return json.Marshal(raw)
}

var _ json.Marshaler = &ServiceAccount{}

type ResourceRecordSets struct {
	Name        string                   `json:"name"`
	ManagedZone string                   `json:"managedZone"`
	Records     []*dns.ResourceRecordSet `json:"records"`
}

func (rrs *ResourceRecordSets) MarshalJSON() ([]byte, error) {
	type NoMethod ResourceRecordSets
	raw := NoMethod(*rrs)
	return json.Marshal(raw)
}

var _ json.Marshaler = &ResourceRecordSets{}

type IAMMemberBinding struct {
	Resource string `json:"resource"`
	Role     string `json:"role"`
	Member   string `json:"member"`
}

func (b *IAMMemberBinding) MarshalJSON() ([]byte, error) {
	type NoMethod IAMMemberBinding
	raw := NoMethod(*b)
	return json.Marshal(raw)
}

var _ json.Marshaler = &IAMMemberBinding{}

// Creates a new deployment with the given name, resources, and labels.
func Create(ctx context.Context, project, name string, resources []Resource, labels map[string]string, dryRun bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.internal.deployments.Create")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("name", name),
		trace.BoolAttribute("dry_run", dryRun),
	)
	defer span.End()

	dm, err := gcp.DeploymentManager(ctx)
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

	return waiter.Poll(ctx, check.DM(ctx, project, op.Name))
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

	dm, err := gcp.DeploymentManager(ctx)
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

	return waiter.Poll(ctx, check.DM(ctx, project, op.Name))
}

// Deletes the given deployment.
func Delete(ctx context.Context, project, name string, dryRun, async bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.internal.deployments.Delete")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("name", name),
		trace.BoolAttribute("dry_run", dryRun),
		trace.BoolAttribute("async", async),
	)
	defer span.End()

	dm, err := gcp.DeploymentManager(ctx)
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

	return waiter.Poll(ctx, check.DM(ctx, project, op.Name))
}

// Lists the deployments for the project, returning the name and labels for each.
func List(ctx context.Context, project string) ([]map[string]string, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.internal.deployments.List")
	span.AddAttributes(
		trace.StringAttribute("project", project),
	)
	defer span.End()

	dm, err := gcp.DeploymentManager(ctx)
	if err != nil {
		return nil, err
	}

	list, err := dm.Deployments.List(project).Do()
	if err != nil {
		return nil, err
	}

	var results []map[string]string
	for _, d := range list.Deployments {
		m := map[string]string{"name": d.Name}
		for _, e := range d.Labels {
			m[e.Key] = e.Value
		}
		results = append(results, m)
	}
	return results, nil
}
