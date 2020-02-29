package belvedere

import (
	"context"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/cfg"
	"github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
	"github.com/codahale/belvedere/pkg/belvedere/internal/gcp"
	"github.com/codahale/belvedere/pkg/belvedere/internal/resources"
	"go.opencensus.io/trace"
)

type AppService interface {
	// List returns a list of apps which have been created in the project.
	List(ctx context.Context) ([]App, error)

	// Create creates an app in the given region with the given name and configuration.
	Create(ctx context.Context, region, name string, config *cfg.Config, dryRun bool, interval time.Duration) error

	// Update updates the resources for the given app to match the given configuration.
	Update(ctx context.Context, name string, config *cfg.Config, dryRun bool, interval time.Duration) error

	// Delete deletes all the resources associated with the given app.
	Delete(ctx context.Context, name string, dryRun, async bool, interval time.Duration) error
}

// App is a Belvedere app.
type App struct {
	Project string
	Region  string
	Name    string
}

type appService struct {
	project string
	dns     *dnsService
	dm      deployments.Manager
}

var _ AppService = &appService{}

func (s *appService) List(ctx context.Context) ([]App, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.apps.List")
	defer span.End()

	// List all deployments in the project.
	list, err := s.dm.List(ctx, s.project, `labels.belvedere-type eq "app"`)
	if err != nil {
		return nil, err
	}

	// Pul app metadata from the labels.
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

	// Validate the app name.
	if err := gcp.ValidateRFC1035(name); err != nil {
		return err
	}

	// Find the project's managed zone.
	managedZone, err := s.dns.findManagedZone(ctx)
	if err != nil {
		return err
	}

	// Create a deployment with all the app resources.
	return s.dm.Insert(ctx, s.project, resources.Name(name),
		resources.App(s.project, name, managedZone, config.CDNPolicy, config.IAP, config.IAMRoles),
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
	managedZone, err := s.dns.findManagedZone(ctx)
	if err != nil {
		return err
	}

	// Update the deployment with the new app resources.
	return s.dm.Update(ctx, s.project, resources.Name(name),
		resources.App(s.project, name, managedZone, config.CDNPolicy, config.IAP, config.IAMRoles),
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

	// Delete the app deployment.
	return s.dm.Delete(ctx, s.project, resources.Name(name), dryRun, async, interval)
}
