package check

import (
	"context"
	"fmt"

	"github.com/codahale/belvedere/pkg/belvedere/internal/gcp"
	"github.com/codahale/belvedere/pkg/belvedere/internal/waiter"
	"go.opencensus.io/trace"
	compute "google.golang.org/api/compute/v0.beta"
)

// Health returns a waiter.Condition for the given instance group being stable and for all its
// instances registering as healthy with the given backend service.
func Health(ctx context.Context, project, region, backendService, instanceGroup string) waiter.Condition {
	return func() (bool, error) {
		ctx, span := trace.StartSpan(ctx, "belvedere.internal.check.Health")
		span.AddAttributes(
			trace.StringAttribute("project", project),
			trace.StringAttribute("region", region),
			trace.StringAttribute("backend_service", backendService),
			trace.StringAttribute("instance_group", instanceGroup),
		)
		defer span.End()

		// Get or create our GCE client.
		gce, err := gcp.Compute(ctx)
		if err != nil {
			return false, err
		}

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
		health, err := gce.BackendServices.GetHealth(project, backendService, &compute.ResourceGroupReference{
			Group: ig.SelfLink,
		}).Context(ctx).Do()
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
