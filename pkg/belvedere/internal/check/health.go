package check

import (
	"context"
	"fmt"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/internal/waiter"
	"go.opencensus.io/trace"
	compute "google.golang.org/api/compute/v1"
)

// HealthChecker provides methods for checking the health of an instance group registered with an
// application's backend service.
type HealthChecker interface {
	Poll(ctx context.Context, project, region, backendService, instanceGroup string, interval time.Duration) error
}

// NewHealthChecker returns a new HealthChecker implementation using the given GCE client.
func NewHealthChecker(gce *compute.Service) HealthChecker {
	return &healthChecker{
		gce: gce,
	}
}

type healthChecker struct {
	gce *compute.Service
}

func (h *healthChecker) Poll(
	ctx context.Context, project, region, backendService, instanceGroup string, interval time.Duration,
) error {
	return waiter.Poll(ctx, interval, Health(ctx, h.gce, project, region, backendService, instanceGroup))
}

// Health returns a waiter.Condition for the given instance group being stable and for all its
// instances registering as healthy with the given backend service.
//nolint:gocognit // this is just complicated
func Health(
	ctx context.Context, gce *compute.Service, project, region, backendService, instanceGroup string,
) waiter.Condition {
	return func() (bool, error) {
		ctx, span := trace.StartSpan(ctx, "belvedere.internal.check.Health")
		defer span.End()

		span.AddAttributes(
			trace.StringAttribute("project", project),
			trace.StringAttribute("region", region),
			trace.StringAttribute("backend_service", backendService),
			trace.StringAttribute("instance_group", instanceGroup),
		)

		// Verify that the instance group manager exists and is stable.
		igm, err := gce.RegionInstanceGroupManagers.Get(project, region, instanceGroup).
			Context(ctx).Fields("status").Do()
		if err != nil {
			return false, fmt.Errorf("getting getting instance group manager: %w", err)
		}

		span.AddAttributes(trace.BoolAttribute("stable", igm.Status.IsStable))

		// If the instance group manager is not stable, continue waiting.
		if !igm.Status.IsStable {
			return false, nil
		}

		// Find the number of running instances.
		ig, err := gce.RegionInstanceGroups.Get(project, region, instanceGroup).
			Context(ctx).Fields("selfLink", "size").Do()
		if err != nil {
			return false, fmt.Errorf("getting getting instance group: %w", err)
		}

		span.AddAttributes(trace.Int64Attribute("instances", ig.Size))

		// Find the health of the running instances.
		health, err := gce.BackendServices.GetHealth(project, backendService,
			&compute.ResourceGroupReference{
				Group: ig.SelfLink,
			},
		).Context(ctx).Do()
		if err != nil {
			return false, fmt.Errorf("getting getting backend service health: %w", err)
		}

		span.AddAttributes(trace.Int64Attribute("registered", int64(len(health.HealthStatus))))

		// If not all instances are registered, continue waiting.
		if len(health.HealthStatus) != int(ig.Size) {
			return false, nil
		}

		// Count the number of healthy instances.
		var healthy int64

		for _, h := range health.HealthStatus {
			span.AddAttributes(trace.StringAttribute("health."+h.Instance, h.HealthState))

			if h.HealthState == "HEALTHY" {
				healthy++
			}
		}

		span.AddAttributes(trace.Int64Attribute("healthy", healthy))

		// If some instances are not healthy, continue waiting.
		return healthy == ig.Size, nil
	}
}
