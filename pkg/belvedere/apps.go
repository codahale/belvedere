package belvedere

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
	"github.com/codahale/belvedere/pkg/belvedere/internal/gcp"
	"github.com/codahale/belvedere/pkg/belvedere/internal/resources"
	"go.opencensus.io/trace"
	"google.golang.org/api/dns/v1"
)

// App is a Belvedere app.
type App struct {
	Project string
	Region  string
	Name    string
}

// Apps returns a list of apps which have been created in the given project.
func Apps(ctx context.Context, project string) ([]App, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.Apps")
	span.AddAttributes(
		trace.StringAttribute("project", project),
	)
	defer span.End()

	// List all deployments in the project.
	list, err := deployments.List(ctx, project, `labels.belvedere-type eq "app"`)
	if err != nil {
		return nil, err
	}

	// Filter the app deployments and pull their metadata from the labels.
	var apps []App
	for _, dep := range list {
		apps = append(apps, App{
			Project: project,
			Name:    dep.App,
			Region:  dep.Region,
		})
	}
	return apps, nil
}

// CreateApp creates an app in the given project and region with the given name and configuration.
func CreateApp(ctx context.Context, project, region, app string, config *Config, dryRun bool, interval time.Duration) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.CreateApp")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("region", region),
		trace.StringAttribute("app", app),
		trace.BoolAttribute("dry_run", dryRun),
	)
	defer span.End()

	// Validate the app name.
	if err := gcp.ValidateRFC1035(app); err != nil {
		return err
	}

	// Find the project's managed zone.
	managedZone, err := findManagedZone(ctx, project)
	if err != nil {
		return err
	}

	// Create a deployment with all the app resources.
	return deployments.Insert(ctx, project, resources.Name(app),
		resources.App(project, app, managedZone, config.CDNPolicy, config.IAP, config.IAMRoles),
		deployments.Labels{
			Type:   "app",
			App:    app,
			Region: region,
		},
		dryRun, interval)
}

// UpdateApp updates the resources for the given app to match the given configuration.
func UpdateApp(ctx context.Context, project, app string, config *Config, dryRun bool, interval time.Duration) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.UpdateApp")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", app),
		trace.BoolAttribute("dry_run", dryRun),
	)
	defer span.End()

	// Find the project's managed zone.
	managedZone, err := findManagedZone(ctx, project)
	if err != nil {
		return err
	}

	// Update the deployment with the new app resources.
	return deployments.Update(ctx, project, resources.Name(app),
		resources.App(project, app, managedZone, config.CDNPolicy, config.IAP, config.IAMRoles),
		dryRun, interval)
}

// DeleteApp deletes all the resources associated with the given app.
func DeleteApp(ctx context.Context, project, app string, dryRun, async bool, interval time.Duration) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.DeleteApp")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", app),
		trace.BoolAttribute("dry_run", dryRun),
		trace.BoolAttribute("async", async),
	)
	defer span.End()

	// Delete the app deployment.
	return deployments.Delete(ctx, project, resources.Name(app), dryRun, async, interval)
}

// findManagedZone returns the Cloud DNS managed zone created via Setup.
func findManagedZone(ctx context.Context, project string) (*dns.ManagedZone, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.findManagedZone")
	span.AddAttributes(
		trace.StringAttribute("project", project),
	)
	defer span.End()

	// Get our DNS client.
	d, err := gcp.DNS(ctx)
	if err != nil {
		return nil, err
	}

	// Find the managed zone.
	mz, err := d.ManagedZones.Get(project, "belvedere").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("error getting managed zone: %w", err)
	}
	return mz, nil
}

// findRegion returns the region the app was created in.
func findRegion(ctx context.Context, project, app string) (string, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.findRegion")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", app),
	)
	defer span.End()

	// Get our DM client.
	dm, err := gcp.DeploymentManager(ctx)
	if err != nil {
		return "", err
	}

	// Find the app deployment.
	deployment, err := dm.Deployments.Get(project, resources.Name(app)).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("error getting deployment: %w", err)
	}

	// Return the app's region given the label.
	for _, l := range deployment.Labels {
		if l.Key == "belvedere-region" {
			return l.Value, nil
		}
	}

	// Handle missing labels.
	return "", errors.New("no region found")
}
