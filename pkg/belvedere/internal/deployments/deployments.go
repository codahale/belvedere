package deployments

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/internal/check"
	"github.com/codahale/belvedere/pkg/belvedere/internal/waiter"
	"go.opencensus.io/trace"
	"google.golang.org/api/deploymentmanager/v2"
	"google.golang.org/api/dns/v1"
	"google.golang.org/api/option"
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

// ServiceAccount represents an IAM service account. This is its own type because Deployment Manager
// doesn't accept the standard API representation.
type ServiceAccount struct {
	AccountID   string `json:"accountId"`
	DisplayName string `json:"displayName"`
}

// MarshalJSON marshals the service account as a JSON object.
func (s *ServiceAccount) MarshalJSON() ([]byte, error) {
	// Cast from a pointer to a raw type to avoid infinite recursion while reusing the standard JSON
	// marshalling code.
	type NoMethod ServiceAccount
	return json.Marshal(NoMethod(*s))
}

var _ json.Marshaler = &ServiceAccount{}

// ResourceRecordSets represents a DNS resource record sets. This is its own type because Deployment
// Manager doesn't accept the standard API representation.
type ResourceRecordSets struct {
	Name        string                   `json:"name"`
	ManagedZone string                   `json:"managedZone"`
	Records     []*dns.ResourceRecordSet `json:"records"`
}

// MarshalJSON marshals the resource record sets as a JSON object.
func (rrs *ResourceRecordSets) MarshalJSON() ([]byte, error) {
	// Cast from a pointer to a raw type to avoid infinite recursion while reusing the standard JSON
	// marshalling code.
	type NoMethod ResourceRecordSets
	return json.Marshal(NoMethod(*rrs))
}

var _ json.Marshaler = &ResourceRecordSets{}

// IAMMemberBinding represents the binding of an IAM role to a project member. This is its own type
// because Deployment Manager doesn't accept the standard API representation.
type IAMMemberBinding struct {
	Resource string `json:"resource"`
	Role     string `json:"role"`
	Member   string `json:"member"`
}

// MarshalJSON marshals the IAM member binding as a JSON object.
func (b *IAMMemberBinding) MarshalJSON() ([]byte, error) {
	// Cast from a pointer to a raw type to avoid infinite recursion while reusing the standard JSON
	// marshalling code.
	type NoMethod IAMMemberBinding
	return json.Marshal(NoMethod(*b))
}

var _ json.Marshaler = &IAMMemberBinding{}

// deploymentConfig is a configuration target for Deployment Manager.
type deploymentConfig struct {
	Resources []Resource `json:"resources"`
}

// Labels are the set of possible deployment labels in use.
type Labels struct {
	Type    string
	Region  string
	App     string
	Release string
	Hash    string
}

// Deployment represents a Belvedere-managed DM deployment.
type Deployment struct {
	Name string
	Labels
}

// Manager provides methods for managing deployments.
type Manager interface {
	// Get returns the deployment with the given name.
	Get(ctx context.Context, project, name string) (*Deployment, error)

	// Insert inserts a new deployment with the given name, resources, and labels.
	Insert(
		ctx context.Context, project, name string, resources []Resource, labels Labels, dryRun bool,
		interval time.Duration,
	) error

	// Update patches the given deployment to add, remove, or modify resources.
	Update(
		ctx context.Context, project, name string, resources []Resource, dryRun bool,
		interval time.Duration,
	) error

	// Delete deletes the given deployment.
	Delete(ctx context.Context, project, name string, dryRun, async bool, interval time.Duration) error

	// List returns a list of deployments in the project which match the given filter.
	List(ctx context.Context, project, filter string) ([]Deployment, error)
}

func NewManager(ctx context.Context, opts ...option.ClientOption) (Manager, error) {
	dm, err := deploymentmanager.NewService(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return &manager{dm: dm}, nil
}

type manager struct {
	dm *deploymentmanager.Service
}

func (m *manager) Get(ctx context.Context, project, name string) (*Deployment, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.internal.deployments.Get")
	defer span.End()

	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("name", name),
	)

	d, err := m.dm.Deployments.Get(project, name).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	return &Deployment{
		Name:   d.Name,
		Labels: entriesToLabels(d.Labels),
	}, nil
}

func (m *manager) Insert(
	ctx context.Context, project, name string, resources []Resource, labels Labels, dryRun bool,
	interval time.Duration,
) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.internal.deployments.Insert")
	defer span.End()

	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("name", name),
		trace.BoolAttribute("dry_run", dryRun),
	)

	// Create our config target.
	d := deploymentConfig{Resources: resources}

	// Pretty-print the config and early exit if we don't want side effects.
	if dryRun {
		b, err := json.MarshalIndent(d, "", "  ")
		if err != nil {
			return fmt.Errorf("error generating JSON: %w", err)
		}

		fmt.Println(string(b))

		return nil
	}

	// Marshal the config target as JSON, since that's parsable by Deployment Manager.
	j, err := json.Marshal(d)
	if err != nil {
		return fmt.Errorf("error generating JSON: %w", err)
	}

	// Insert the new deployment.
	op, err := m.dm.Deployments.Insert(project, &deploymentmanager.Deployment{
		Labels: labelsToEntries(&labels),
		Name:   name,
		Target: &deploymentmanager.TargetConfiguration{
			Config: &deploymentmanager.ConfigFile{
				Content: string(j),
			},
		},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("error inserting deployment: %w", err)
	}

	// Wait for the deployment to be created or fail.
	return waiter.Poll(ctx, interval, check.DM(ctx, m.dm, project, op.Name))
}

func (m *manager) Update(
	ctx context.Context, project, name string, resources []Resource, dryRun bool,
	interval time.Duration,
) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.internal.deployments.Update")
	defer span.End()

	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("name", name),
		trace.BoolAttribute("dry_run", dryRun),
	)

	// Create our config target.
	d := deploymentConfig{Resources: resources}

	// Pretty-print the config and early exit if we don't want side effects.
	if dryRun {
		b, err := json.MarshalIndent(d, "", "  ")
		if err != nil {
			return fmt.Errorf("error generating JSON: %w", err)
		}

		fmt.Println(string(b))

		return nil
	}

	// Marshal the config target as JSON, since that's parsable by Deployment Manager.
	j, err := json.Marshal(d)
	if err != nil {
		return fmt.Errorf("error generating JSON: %w", err)
	}

	// Update the deployment.
	op, err := m.dm.Deployments.Patch(project, name, &deploymentmanager.Deployment{
		Target: &deploymentmanager.TargetConfiguration{
			Config: &deploymentmanager.ConfigFile{
				Content: string(j),
			},
		},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("error patching deployment: %w", err)
	}

	// Wait for the deployment to be updated or fail.
	return waiter.Poll(ctx, interval, check.DM(ctx, m.dm, project, op.Name))
}

func (m *manager) Delete(ctx context.Context, project, name string, dryRun, async bool, interval time.Duration) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.internal.deployments.Delete")
	defer span.End()

	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("name", name),
		trace.BoolAttribute("dry_run", dryRun),
		trace.BoolAttribute("async", async),
	)

	// Early exit if we don't want side effects.
	if dryRun {
		return nil
	}

	// Delete the deployment.
	op, err := m.dm.Deployments.Delete(project, name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("error deleting deployment: %w", err)
	}

	// Early exit if we don't care about results.
	if async {
		return nil
	}

	// Wait for the deployment to be deleted or fail.
	return waiter.Poll(ctx, interval, check.DM(ctx, m.dm, project, op.Name))
}

func (m *manager) List(ctx context.Context, project, filter string) ([]Deployment, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.internal.deployments.List")
	defer span.End()

	span.AddAttributes(
		trace.StringAttribute("project", project),
	)

	var deployments []Deployment

	// List all of the deployments.
	if err := m.dm.Deployments.List(project).Filter(filter).Pages(ctx,
		func(list *deploymentmanager.DeploymentsListResponse) error {
			// Convert labels to maps.
			for _, d := range list.Deployments {
				deployments = append(deployments, Deployment{
					Name:   d.Name,
					Labels: entriesToLabels(d.Labels),
				})
			}
			return nil
		},
	); err != nil {
		return nil, fmt.Errorf("error listing deployments: %w", err)
	}

	sort.SliceStable(deployments, func(i, j int) bool {
		return deployments[i].Name < deployments[j].Name
	})

	return deployments, nil
}

func entry(k, v string) *deploymentmanager.DeploymentLabelEntry {
	return &deploymentmanager.DeploymentLabelEntry{Key: k, Value: v}
}

func labelsToEntries(l *Labels) []*deploymentmanager.DeploymentLabelEntry {
	var entries []*deploymentmanager.DeploymentLabelEntry
	if l.App != "" {
		entries = append(entries, entry("belvedere-app", l.App))
	}

	if l.Hash != "" {
		entries = append(entries, entry("belvedere-hash", l.Hash))
	}

	if l.Region != "" {
		entries = append(entries, entry("belvedere-region", l.Region))
	}

	if l.Release != "" {
		entries = append(entries, entry("belvedere-release", l.Release))
	}

	entries = append(entries, entry("belvedere-type", l.Type))

	return entries
}

func entriesToLabels(entries []*deploymentmanager.DeploymentLabelEntry) Labels {
	var l Labels

	for _, e := range entries {
		switch e.Key {
		case "belvedere-app":
			l.App = e.Value
		case "belvedere-hash":
			l.Hash = e.Value
		case "belvedere-region":
			l.Region = e.Value
		case "belvedere-release":
			l.Release = e.Value
		case "belvedere-type":
			l.Type = e.Value
		}
	}

	return l
}
