package belvedere

import (
	"context"
	"strings"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
	"github.com/codahale/belvedere/pkg/belvedere/internal/resources"
	"github.com/codahale/belvedere/pkg/belvedere/internal/setup"
	"go.opencensus.io/trace"
)

// Setup enables all required GCP services, grants Deployment Manager the permissions required to
// manage service accounts and IAM roles, and creates a deployment with the base resources needed
// to use Belvedere.
func Setup(ctx context.Context, project, dnsZone string, dryRun bool, interval time.Duration) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.Setup")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("dns_zone", dnsZone),
		trace.BoolAttribute("dry_run", dryRun),
	)
	defer span.End()

	// Enable all required services.
	if err := setup.EnableAPIs(ctx, project, dryRun, interval); err != nil {
		return err
	}

	// Grant Deployment Manager the required permissions to manage IAM roles.
	if err := setup.SetDMPerms(ctx, project, dryRun); err != nil {
		return err
	}

	// Ensure the DNS zone ends with a period.
	if !strings.HasSuffix(dnsZone, ".") {
		dnsZone += "."
	}

	// Create a deployment with a managed DNS zone and firewall rules which limit SSH to GCE
	// instances to those tunneled over IAP.
	return deployments.Insert(ctx, project, "belvedere", resources.Base(dnsZone),
		map[string]string{
			"belvedere-type": "base",
		}, dryRun, interval)
}

// Teardown deletes the shared firewall rules and managed zone created by Setup. It does not disable
// services or downgrade Deployment Manager's permissions.
func Teardown(ctx context.Context, project string, dryRun, async bool, interval time.Duration) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.Teardown")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.BoolAttribute("dry_run", dryRun),
		trace.BoolAttribute("async", async),
	)
	defer span.End()

	// Delete the shared deployment.
	return deployments.Delete(ctx, project, "belvedere", dryRun, async, interval)
}
