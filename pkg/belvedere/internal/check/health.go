package check

import (
	"context"

	"go.opencensus.io/trace"
	"google.golang.org/api/compute/v0.beta"
	"k8s.io/apimachinery/pkg/util/wait"
)

func Health(ctx context.Context, gce *compute.Service, project, region, backendService, instanceGroup string) wait.ConditionFunc {
	return func() (bool, error) {
		ctx, span := trace.StartSpan(ctx, "belvedere.internal.check.Health")
		span.AddAttributes(trace.StringAttribute("project", project))
		span.AddAttributes(trace.StringAttribute("region", region))
		span.AddAttributes(trace.StringAttribute("backend_service", backendService))
		span.AddAttributes(trace.StringAttribute("instance_group", instanceGroup))
		defer span.End()

		// Verify that the instance group manager exists and is stable.
		igm, err := gce.RegionInstanceGroupManagers.Get(project, region, instanceGroup).
			Context(ctx).Fields("status").Do()
		if err != nil {
			return false, err
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
			return false, err
		}
		span.AddAttributes(trace.Int64Attribute("instances", ig.Size))

		// Find the health of the running instances.
		health, err := gce.BackendServices.GetHealth(project, backendService, &compute.ResourceGroupReference{
			Group: ig.SelfLink,
		}).Context(ctx).Do()
		if err != nil {
			return false, err
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
