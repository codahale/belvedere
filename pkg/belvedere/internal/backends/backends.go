package backends

import (
	"context"

	"github.com/codahale/belvedere/pkg/belvedere/internal/check"
	"github.com/codahale/belvedere/pkg/belvedere/internal/gcp"
	"github.com/codahale/belvedere/pkg/belvedere/internal/waiter"
	"go.opencensus.io/trace"
	"google.golang.org/api/compute/v0.beta"
)

// Adds an instance group to a backend service.
func Add(ctx context.Context, project, region, backendService, instanceGroup string, dryRun bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.internal.backends.Add")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("region", region),
		trace.StringAttribute("backend_service", backendService),
		trace.StringAttribute("instance_group", instanceGroup),
		trace.BoolAttribute("dry_run", dryRun),
	)
	defer span.End()

	gce, err := gcp.Compute(ctx)
	if err != nil {
		return err
	}

	// Get the current backends.
	bes, err := gce.BackendServices.Get(project, backendService).
		Context(ctx).Fields("backends", "fingerprint").Do()
	if err != nil {
		return err
	}

	// Get the instance group's full URL.
	ig, err := gce.RegionInstanceGroups.Get(project, region, instanceGroup).
		Context(ctx).Fields("selfLink").Do()
	if err != nil {
		return err
	}

	// Check to see if the instance group is already in service.
	for _, be := range bes.Backends {
		if be.Group == ig.SelfLink {
			span.AddAttributes(trace.BoolAttribute("modified", false))
			return nil
		}
	}
	span.AddAttributes(trace.BoolAttribute("modified", true))

	// Early exit if we don't want side effects.
	if dryRun {
		return nil
	}

	// Patch the backend service to include the instance group as a backend.
	op, err := gce.BackendServices.Patch(project, backendService,
		&compute.BackendService{
			Backends: append(bes.Backends, &compute.Backend{
				Group: ig.SelfLink,
			}),
			// Include the fingerprint to avoid overwriting concurrent writes.
			Fingerprint:     bes.Fingerprint,
			ForceSendFields: []string{"Backends", "Fingerprint"},
		},
	).Context(ctx).Do()
	if err != nil {
		return err
	}

	// Return patch operation.
	return waiter.Poll(ctx, check.GCE(ctx, project, op.Name))
}

// Removes an instance group from a backend service.
func Remove(ctx context.Context, project, region, backendService, instanceGroup string, dryRun bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.internal.backends.Remove")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("region", region),
		trace.StringAttribute("backend_service", backendService),
		trace.StringAttribute("instance_group", instanceGroup),
		trace.BoolAttribute("dry_run", dryRun),
	)
	defer span.End()

	gce, err := gcp.Compute(ctx)
	if err != nil {
		return err
	}

	// Get the current backends.
	bes, err := gce.BackendServices.Get(project, backendService).
		Context(ctx).Fields("backends", "fingerprint").Do()
	if err != nil {
		return err
	}

	// Get the instance group's full URL.
	ig, err := gce.RegionInstanceGroups.Get(project, region, instanceGroup).
		Context(ctx).Fields("selfLink").Do()
	if err != nil {
		return err
	}

	// Copy all backends except for the instance group in question.
	var backends []*compute.Backend
	for _, be := range bes.Backends {
		if be.Group != ig.SelfLink {
			backends = append(backends, be)
		}
	}

	// Early exit if the instance group isn't in service and doesn't need to be removed.
	if len(bes.Backends) == len(backends) {
		span.AddAttributes(trace.BoolAttribute("modified", false))
		return nil
	}
	span.AddAttributes(trace.BoolAttribute("modified", true))

	// Early exit if we don't want side effects.
	if dryRun {
		return nil
	}

	// Patch the backend service to remove the instance group as a backend.
	op, err := gce.BackendServices.Patch(project, backendService,
		&compute.BackendService{
			Backends: backends,
			// Include the fingerprint to avoid overwriting concurrent writes.
			Fingerprint: bes.Fingerprint,
			// Force sending both fields in case the backends list is empty.
			ForceSendFields: []string{"Backends", "Fingerprint"},
		},
	).Context(ctx).Do()
	if err != nil {
		return err
	}

	// Return the patch operation.
	return waiter.Poll(ctx, check.GCE(ctx, project, op.Name))
}
