package belvedere

import (
	"context"
	"fmt"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/internal/backends"
	"github.com/codahale/belvedere/pkg/belvedere/internal/check"
	"github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
	"github.com/codahale/belvedere/pkg/belvedere/internal/gcp"
	"github.com/codahale/belvedere/pkg/belvedere/internal/resources"
	"github.com/codahale/belvedere/pkg/belvedere/internal/waiter"
	"go.opencensus.io/trace"
)

// A Release describes a specific release of an app.
type Release struct {
	Project string
	Region  string
	App     string
	Release string
	Hash    string
}

// Releases returns a list of releases in the given project for the given app, if any is passed.
func Releases(ctx context.Context, project, app string) ([]Release, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.Releases")
	span.AddAttributes(
		trace.StringAttribute("project", project),
	)
	defer span.End()

	if app != "" {
		span.AddAttributes(trace.StringAttribute("app", app))
	}

	filter := `labels.belvedere-type eq "release"`
	if app != "" {
		filter = fmt.Sprintf("%s AND labels.belvedere-app eq %q", filter, app)
	}

	list, err := deployments.List(ctx, project, filter)
	if err != nil {
		return nil, err
	}

	var releases []Release
	for _, dep := range list {
		releases = append(releases, Release{
			Project: project,
			Region:  dep.Region,
			App:     dep.App,
			Release: dep.Release,
			Hash:    dep.Hash,
		})
	}
	return releases, nil
}

// CreateRelease creates a deployment containing release resources for the given app.
func CreateRelease(ctx context.Context, project, app, release string, config *Config, imageSHA256 string, dryRun bool, interval time.Duration) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.CreateRelease")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", app),
		trace.StringAttribute("release", release),
		trace.StringAttribute("image_sha256", imageSHA256),
		trace.BoolAttribute("dry_run", dryRun),
	)
	defer span.End()

	if err := gcp.ValidateRFC1035(release); err != nil {
		return err
	}

	region, err := findRegion(ctx, project, app)
	if err != nil {
		return err
	}

	return deployments.Insert(ctx, project, resources.Name(app, release),
		resources.Release(
			project, region, app, release, config.Network, config.Subnetwork,
			config.MachineType, config.cloudConfig(app, release, imageSHA256),
			config.NumReplicas, config.AutoscalingPolicy,
		),
		deployments.Labels{
			Type:    "release",
			App:     app,
			Release: release,
			Region:  region,
			Hash:    imageSHA256[:32],
		}, dryRun, interval)
}

// EnableRelease adds the release's instance group to the app's backend service and waits for the
// instances to go fully into service.
func EnableRelease(ctx context.Context, project, app, release string, dryRun bool, interval time.Duration) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.EnableRelease")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", app),
		trace.StringAttribute("release", release),
		trace.BoolAttribute("dry_run", dryRun),
	)
	defer span.End()

	region, err := findRegion(ctx, project, app)
	if err != nil {
		return err
	}

	backendService := fmt.Sprintf("%s-bes", app)
	instanceGroup := fmt.Sprintf("%s-%s-ig", app, release)
	if err := backends.Add(ctx, project, region, backendService, instanceGroup, dryRun, interval); err != nil {
		return err
	}

	return waiter.Poll(ctx, interval, check.Health(ctx, project, region, backendService, instanceGroup))
}

// DisableRelease removes the release's instance group from the app's backend service.
func DisableRelease(ctx context.Context, project, app, release string, dryRun bool, interval time.Duration) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.DisableRelease")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", app),
		trace.StringAttribute("release", release),
		trace.BoolAttribute("dry_run", dryRun),
	)
	defer span.End()

	region, err := findRegion(ctx, project, app)
	if err != nil {
		return err
	}

	backendService := fmt.Sprintf("%s-bes", app)
	instanceGroup := fmt.Sprintf("%s-%s-ig", app, release)
	return backends.Remove(ctx, project, region, backendService, instanceGroup, dryRun, interval)
}

// DeleteRelease deletes the release's deployment and waits for all underlying resources to be
// deleted.
func DeleteRelease(ctx context.Context, project, app, release string, dryRun, async bool, interval time.Duration) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.DeleteRelease")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", app),
		trace.StringAttribute("release", release),
		trace.BoolAttribute("dry_run", dryRun),
		trace.BoolAttribute("async", async),
	)
	defer span.End()

	return deployments.Delete(ctx, project, resources.Name(app, release), dryRun, async, interval)
}
