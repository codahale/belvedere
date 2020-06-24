package belvedere

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/cfg"
	"github.com/codahale/belvedere/pkg/belvedere/internal/backends"
	"github.com/codahale/belvedere/pkg/belvedere/internal/check"
	"github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
	"github.com/codahale/belvedere/pkg/belvedere/internal/gcp"
	"github.com/codahale/belvedere/pkg/belvedere/internal/resources"
	"go.opencensus.io/trace"
	compute "google.golang.org/api/compute/v1"
)

// A Release describes a specific release of an app.
type Release struct {
	Project string
	Region  string
	App     string
	Release string
	Hash    string
}

// ReleaseService provides methods for managing releases.
type ReleaseService interface {
	// List returns a list of releases in the given project for the given app, if any is passed.
	List(ctx context.Context, app string) ([]Release, error)

	// Create creates a deployment containing release resources for the given app.
	Create(
		ctx context.Context, app, name string, config *cfg.Config, imageSHA256 string, dryRun bool,
		interval time.Duration,
	) error

	// Enable adds the release's instance group to the app's backend project and waits for the
	// instances to go fully into project.
	Enable(ctx context.Context, app, name string, dryRun bool, interval time.Duration) error

	// Disable removes the release's instance group from the app's backend project.
	Disable(ctx context.Context, app, name string, dryRun bool, interval time.Duration) error

	// Delete deletes the release's deployment and waits for all underlying resources to be deleted.
	Delete(ctx context.Context, app, name string, dryRun, async bool, interval time.Duration) error
}

type releaseService struct {
	project   string
	dm        deployments.Manager
	gce       *compute.Service
	backends  backends.Service
	resources resources.Builder
	health    check.HealthChecker
	apps      AppService
}

func (r *releaseService) List(ctx context.Context, app string) ([]Release, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.releases.List")
	defer span.End()

	if app != "" {
		span.AddAttributes(trace.StringAttribute("app", app))
	}

	filter := `labels.belvedere-type eq "release"`
	if app != "" {
		filter = fmt.Sprintf("%s AND labels.belvedere-app eq %q", filter, app)
	}

	list, err := r.dm.List(ctx, r.project, filter)
	if err != nil {
		return nil, err
	}

	releases := make([]Release, len(list))
	for i, dep := range list {
		releases[i] = Release{
			Project: r.project,
			Region:  dep.Region,
			App:     dep.App,
			Release: dep.Release,
			Hash:    dep.Hash,
		}
	}

	return releases, nil
}

// nolint:gochecknoglobals
var imageHashFormat = regexp.MustCompile(`^[a-f0-9]{64}$`)

type InvalidSHA256DigestError struct {
	Digest string
}

func (e *InvalidSHA256DigestError) Error() string {
	return fmt.Sprintf("invalid SHA-256 digest: %q", e.Digest)
}

func (r *releaseService) Create(
	ctx context.Context, app, name string, config *cfg.Config, imageSHA256 string, dryRun bool,
	interval time.Duration,
) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.releases.Create")
	defer span.End()

	span.AddAttributes(
		trace.StringAttribute("app", app),
		trace.StringAttribute("name", name),
		trace.StringAttribute("image_sha256", imageSHA256),
		trace.BoolAttribute("dry_run", dryRun),
	)

	if err := gcp.ValidateRFC1035(name); err != nil {
		return err
	}

	if !imageHashFormat.MatchString(imageSHA256) {
		return &InvalidSHA256DigestError{Digest: imageSHA256}
	}

	a, err := r.apps.Get(ctx, app)
	if err != nil {
		return err
	}

	return r.dm.Insert(ctx, r.project, resources.Name(app, name),
		r.resources.Release(r.project, a.Region, app, name, imageSHA256, config),
		deployments.Labels{
			Type:    "release",
			App:     app,
			Release: name,
			Region:  a.Region,
			Hash:    imageSHA256[:32],
		},
		dryRun, interval,
	)
}

func (r *releaseService) Enable(ctx context.Context, app, name string, dryRun bool, interval time.Duration) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.releases.Enable")
	defer span.End()

	span.AddAttributes(
		trace.StringAttribute("app", app),
		trace.StringAttribute("name", name),
		trace.BoolAttribute("dry_run", dryRun),
	)

	a, err := r.apps.Get(ctx, app)
	if err != nil {
		return err
	}

	backendService := fmt.Sprintf("%s-bes", app)
	instanceGroup := fmt.Sprintf("%s-%s-ig", app, name)

	if err := r.backends.Add(ctx, r.project, a.Region, backendService, instanceGroup, dryRun, interval); err != nil {
		return err
	}

	return r.health.Poll(ctx, r.project, a.Region, backendService, instanceGroup, interval)
}

func (r *releaseService) Disable(ctx context.Context, app, name string, dryRun bool, interval time.Duration) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.releases.Disable")
	defer span.End()

	span.AddAttributes(
		trace.StringAttribute("app", app),
		trace.StringAttribute("name", name),
		trace.BoolAttribute("dry_run", dryRun),
	)

	a, err := r.apps.Get(ctx, app)
	if err != nil {
		return err
	}

	backendService := fmt.Sprintf("%s-bes", app)
	instanceGroup := fmt.Sprintf("%s-%s-ig", app, name)

	return r.backends.Remove(ctx, r.project, a.Region, backendService, instanceGroup, dryRun, interval)
}

func (r *releaseService) Delete(
	ctx context.Context, app, name string, dryRun, async bool, interval time.Duration,
) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.releases.Delete")
	defer span.End()

	span.AddAttributes(
		trace.StringAttribute("app", app),
		trace.StringAttribute("name", name),
		trace.BoolAttribute("dry_run", dryRun),
		trace.BoolAttribute("async", async),
	)

	return r.dm.Delete(ctx, r.project, resources.Name(app, name), dryRun, async, interval)
}
