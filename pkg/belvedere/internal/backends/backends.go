package backends

import (
	"context"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/internal/check"
	"go.opencensus.io/trace"
	"google.golang.org/api/compute/v0.beta"
	"k8s.io/apimachinery/pkg/util/wait"
)

func Add(ctx context.Context, gce *compute.Service, project, region, backendService, instanceGroup string, dryRun bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.internal.backends.Add")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("region", region),
		trace.StringAttribute("backend_service", backendService),
		trace.StringAttribute("instance_group", instanceGroup),
		trace.BoolAttribute("dry_run", dryRun),
	)
	defer span.End()

	bes, err := gce.BackendServices.Get(project, backendService).
		Context(ctx).Fields("backends", "fingerprint").Do()
	if err != nil {
		return err
	}

	ig, err := gce.RegionInstanceGroups.Get(project, region, instanceGroup).
		Context(ctx).Fields("selfLink").Do()
	if err != nil {
		return err
	}

	for _, be := range bes.Backends {
		if be.Group == ig.SelfLink {
			span.AddAttributes(trace.BoolAttribute("modified", false))
			return nil
		}
	}
	bes.Backends = append(bes.Backends, &compute.Backend{
		Group: ig.SelfLink,
	})

	if dryRun {
		return nil
	}

	op, err := gce.BackendServices.Patch(project, backendService,
		&compute.BackendService{
			Backends:        bes.Backends,
			Fingerprint:     bes.Fingerprint,
			ForceSendFields: []string{"Backends", "Fingerprint"},
		},
	).Context(ctx).Do()
	if err != nil {
		return err
	}

	span.AddAttributes(trace.BoolAttribute("modified", true))
	return wait.Poll(10*time.Second, 5*time.Minute, check.GCE(ctx, gce, project, op.Name))
}

func Remove(ctx context.Context, gce *compute.Service, project, region, backendService, instanceGroup string, dryRun bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.internal.backends.Remove")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("region", region),
		trace.StringAttribute("backend_service", backendService),
		trace.StringAttribute("instance_group", instanceGroup),
		trace.BoolAttribute("dry_run", dryRun),
	)
	defer span.End()

	ig, err := gce.RegionInstanceGroups.Get(project, region, instanceGroup).
		Context(ctx).Fields("selfLink").Do()
	if err != nil {
		return err
	}

	bes, err := gce.BackendServices.Get(project, backendService).
		Context(ctx).Fields("backends", "fingerprint").Do()
	if err != nil {
		return err
	}

	var backends []*compute.Backend
	for _, be := range bes.Backends {
		if be.Group != ig.SelfLink {
			backends = append(backends, be)
		}
	}

	if len(bes.Backends) == len(backends) {
		span.AddAttributes(trace.BoolAttribute("modified", false))
		return nil
	}

	if dryRun {
		return nil
	}

	op, err := gce.BackendServices.Patch(project, backendService,
		&compute.BackendService{
			Backends:        backends,
			Fingerprint:     bes.Fingerprint,
			ForceSendFields: []string{"Backends", "Fingerprint"},
		},
	).Context(ctx).Do()
	if err != nil {
		return err
	}

	span.AddAttributes(trace.BoolAttribute("modified", true))
	return wait.Poll(10*time.Second, 5*time.Minute, check.GCE(ctx, gce, project, op.Name))
}
