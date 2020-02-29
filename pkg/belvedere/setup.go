package belvedere

import (
	"context"
	"strings"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
	"github.com/codahale/belvedere/pkg/belvedere/internal/resources"
	"go.opencensus.io/trace"
)

func (p *project) Setup(ctx context.Context, dnsZone string, dryRun bool, interval time.Duration) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.project.Setup")
	span.AddAttributes(
		trace.StringAttribute("dns_zone", dnsZone),
		trace.BoolAttribute("dry_run", dryRun),
	)
	defer span.End()

	// Enable all required services.
	if err := p.setup.EnableAPIs(ctx, p.name, dryRun, interval); err != nil {
		return err
	}

	// Grant Deployment Manager the required permissions to manage IAM roles.
	if err := p.setup.SetDMPerms(ctx, p.name, dryRun); err != nil {
		return err
	}

	// Ensure the DNS zone ends with a period.
	if !strings.HasSuffix(dnsZone, ".") {
		dnsZone += "."
	}

	// Create a deployment with a managed DNS zone and firewall rules which limit SSH to GCE
	// instances to those tunneled over IAP.
	return p.dm.Insert(ctx, p.name, "belvedere", resources.Base(dnsZone),
		deployments.Labels{
			Type: "base",
		},
		dryRun, interval,
	)
}

func (p *project) Teardown(ctx context.Context, dryRun, async bool, interval time.Duration) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.project.Teardown")
	span.AddAttributes(
		trace.BoolAttribute("dry_run", dryRun),
		trace.BoolAttribute("async", async),
	)
	defer span.End()

	// Delete the shared deployment.
	return p.dm.Delete(ctx, p.name, "belvedere", dryRun, async, interval)
}
