package backends

import (
	"context"
	"fmt"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/internal/check"
	"github.com/codahale/belvedere/pkg/belvedere/internal/gcp"
	"github.com/codahale/belvedere/pkg/belvedere/internal/waiter"
	"go.opencensus.io/trace"
	compute "google.golang.org/api/compute/v1"
)

// Service provides methods for managing the backend services associated with an application's
// load balancer.
type Service interface {
	// Add adds an instance group to a backend service. If the instance group is already registered as a
	// backend, exits early.
	Add(ctx context.Context, project, region, backendService, instanceGroup string,
		dryRun bool, interval time.Duration) error

	// Remove removes an instance group from a backend service. If the instance group is not registered
	// as a backend, exits early.
	Remove(ctx context.Context, project, region, backendService, instanceGroup string,
		dryRun bool, interval time.Duration) error
}

// NewService returns a new Service implementation.
func NewService(gce *compute.Service) Service {
	return &service{gce: gce}
}

type service struct {
	gce *compute.Service
}

// nolint: gocognit
func (s *service) Add(ctx context.Context, project, region, backendService, instanceGroup string,
	dryRun bool, interval time.Duration) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.internal.backends.Add")
	defer span.End()

	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("region", region),
		trace.StringAttribute("backend_service", backendService),
		trace.StringAttribute("instance_group", instanceGroup),
		trace.BoolAttribute("dry_run", dryRun),
	)

	var op *compute.Operation

	err := gcp.ModifyLoop(5*time.Second, 2*time.Minute, func() error {
		// Get the current backends.
		bes, err := s.gce.BackendServices.Get(project, backendService).
			Context(ctx).Fields("backends", "fingerprint").Do()
		if err != nil {
			return fmt.Errorf("error getting backend service: %w", err)
		}

		// Get the instance group's full URL.
		ig, err := s.gce.RegionInstanceGroups.Get(project, region, instanceGroup).
			Context(ctx).Fields("selfLink").Do()
		if err != nil {
			return fmt.Errorf("error getting instance group: %w", err)
		}

		// Check to see if the instance group is already in service.
		idx := findBackend(bes.Backends, ig.SelfLink)
		if idx >= 0 {
			span.AddAttributes(trace.BoolAttribute("modified", false))
			return nil
		}
		span.AddAttributes(trace.BoolAttribute("modified", true))

		// Early exit if we don't want side effects.
		if dryRun {
			return nil
		}

		// Patch the backend service to include the instance group as a backend.
		op, err = s.gce.BackendServices.Patch(project, backendService,
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
			return fmt.Errorf("error patching backend service: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Continue the early exit, if necessary.
	if op == nil {
		return nil
	}

	// Return patch operation.
	return waiter.Poll(ctx, interval, check.GCE(ctx, s.gce, project, op.Name))
}

// nolint: gocognit
func (s *service) Remove(
	ctx context.Context, project, region, backendService, instanceGroup string, dryRun bool,
	interval time.Duration,
) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.internal.backends.Remove")
	defer span.End()

	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("region", region),
		trace.StringAttribute("backend_service", backendService),
		trace.StringAttribute("instance_group", instanceGroup),
		trace.BoolAttribute("dry_run", dryRun),
	)

	var op *compute.Operation

	err := gcp.ModifyLoop(5*time.Second, 2*time.Minute, func() error {
		// Get the current backends.
		bes, err := s.gce.BackendServices.Get(project, backendService).
			Context(ctx).Fields("backends", "fingerprint").Do()
		if err != nil {
			return fmt.Errorf("error getting backend service: %w", err)
		}

		// Get the instance group's full URL.
		ig, err := s.gce.RegionInstanceGroups.Get(project, region, instanceGroup).
			Context(ctx).Fields("selfLink").Do()
		if err != nil {
			return fmt.Errorf("error getting instance group: %w", err)
		}

		// Find the index of the instance group.
		idx := findBackend(bes.Backends, ig.SelfLink)
		if idx < 0 {
			// Early exit if the instance group isn't in service and doesn't need to be removed.
			span.AddAttributes(trace.BoolAttribute("modified", false))
			return nil
		}

		// Remove the instance group in question from the backends.
		bes.Backends = append(bes.Backends[:idx], bes.Backends[idx+1:]...)
		span.AddAttributes(trace.BoolAttribute("modified", true))

		// Early exit if we don't want side effects.
		if dryRun {
			return nil
		}

		// Patch the backend service to remove the instance group as a backend.
		op, err = s.gce.BackendServices.Patch(project, backendService,
			&compute.BackendService{
				Backends: bes.Backends,
				// Include the fingerprint to avoid overwriting concurrent writes.
				Fingerprint: bes.Fingerprint,
				// Force sending both fields in case the backends list is empty.
				ForceSendFields: []string{"Backends", "Fingerprint"},
			},
		).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("error patching backend service: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Continue the early exit, if necessary.
	if op == nil {
		return nil
	}

	// Return the patch operation.
	return waiter.Poll(ctx, interval, check.GCE(ctx, s.gce, project, op.Name))
}

func findBackend(backends []*compute.Backend, ig string) int {
	for i, be := range backends {
		if be.Group == ig {
			return i
		}
	}

	return -1
}
