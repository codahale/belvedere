package belvedere

import (
	"context"
	"fmt"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/cfg"
	"github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
	"github.com/codahale/belvedere/pkg/belvedere/internal/gcp"
	"github.com/codahale/belvedere/pkg/belvedere/internal/resources"
	"github.com/codahale/belvedere/pkg/belvedere/internal/setup"
	"go.opencensus.io/trace"
	compute "google.golang.org/api/compute/v0.beta"
)

// AppService provides methods for managing applications.
type AppService interface {
	// Get returns the application with the given name.
	Get(ctx context.Context, name string) (*App, error)

	// List returns a list of applications which have been created in the project.
	List(ctx context.Context) ([]App, error)

	// Create creates an application in the given region with the given name and configuration.
	Create(ctx context.Context, region, name string, config *cfg.Config, dryRun bool, interval time.Duration) error

	// Update updates the resources for the given application to match the given configuration.
	Update(ctx context.Context, name string, config *cfg.Config, dryRun bool, interval time.Duration) error

	// Delete deletes all the resources associated with the given application.
	Delete(ctx context.Context, name string, dryRun, async bool, interval time.Duration) error
}

// App is a Belvedere application.
type App struct {
	Project string
	Region  string
	Name    string
}

type appService struct {
	project   string
	setup     setup.Service
	dm        deployments.Manager
	resources resources.Builder
	gce       *compute.Service
}

var _ AppService = &appService{}

func (s *appService) Get(ctx context.Context, name string) (*App, error) {
	dep, err := s.dm.Get(ctx, s.project, resources.Name(name))
	if err != nil {
		return nil, err
	}
	return &App{
		Project: s.project,
		Name:    dep.App,
		Region:  dep.Region,
	}, nil
}

func (s *appService) List(ctx context.Context) ([]App, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.apps.List")
	defer span.End()

	// List all deployments in the project.
	list, err := s.dm.List(ctx, s.project, `labels.belvedere-type eq "app"`)
	if err != nil {
		return nil, err
	}

	// Pul application metadata from the labels.
	apps := make([]App, len(list))
	for i, dep := range list {
		apps[i] = App{
			Project: s.project,
			Name:    dep.App,
			Region:  dep.Region,
		}
	}
	return apps, nil
}

func (s *appService) Create(ctx context.Context, region, name string, config *cfg.Config, dryRun bool, interval time.Duration) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.apps.Create")
	span.AddAttributes(
		trace.StringAttribute("region", region),
		trace.StringAttribute("name", name),
		trace.BoolAttribute("dry_run", dryRun),
	)
	defer span.End()

	// Validate the application name.
	if err := gcp.ValidateRFC1035(name); err != nil {
		return err
	}

	// Validate the region name and status.
	r, err := s.gce.Regions.Get(s.project, region).Context(ctx).Fields("status").Do()
	if err != nil {
		return fmt.Errorf("invalid region %q: %w", region, err)
	}
	if r.Status != "UP" {
		return fmt.Errorf("region %q is down", region)
	}

	// Find the project's managed zone.
	managedZone, err := s.setup.ManagedZone(ctx, s.project)
	if err != nil {
		return err
	}

	// Create a deployment with all the application resources.
	return s.dm.Insert(ctx, s.project, resources.Name(name),
		s.resources.App(s.project, name, managedZone, config),
		deployments.Labels{
			Type:   "app",
			App:    name,
			Region: region,
		},
		dryRun, interval,
	)
}

func (s *appService) Update(ctx context.Context, name string, config *cfg.Config, dryRun bool, interval time.Duration) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.apps.Update")
	span.AddAttributes(
		trace.StringAttribute("name", name),
		trace.BoolAttribute("dry_run", dryRun),
	)
	defer span.End()

	// Find the project's managed zone.
	managedZone, err := s.setup.ManagedZone(ctx, s.project)
	if err != nil {
		return err
	}

	// Update the deployment with the new application resources.
	return s.dm.Update(ctx, s.project, resources.Name(name),
		s.resources.App(s.project, name, managedZone, config),
		dryRun, interval,
	)
}

func (s *appService) Delete(ctx context.Context, name string, dryRun, async bool, interval time.Duration) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.apps.Delete")
	span.AddAttributes(
		trace.StringAttribute("name", name),
		trace.BoolAttribute("dry_run", dryRun),
		trace.BoolAttribute("async", async),
	)
	defer span.End()

	// Delete the application deployment.
	return s.dm.Delete(ctx, s.project, resources.Name(name), dryRun, async, interval)
}
