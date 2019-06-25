package belvedere

import (
	"context"
	"fmt"
	"strings"

	"github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
	"github.com/codahale/belvedere/pkg/belvedere/internal/setup"
	"go.opencensus.io/trace"
	"google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/dns/v1"
)

// Setup enables all required GCP services, grants Deployment Manager the permissions required to
// manage service accounts and IAM roles, and creates a deployment with the base resources needed
// to use Belvedere.
func Setup(ctx context.Context, project, dnsZone string, dryRun bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.Setup")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("dns_zone", dnsZone),
		trace.BoolAttribute("dry_run", dryRun),
	)
	defer span.End()

	// Enable all required services.
	if err := setup.EnableAPIs(ctx, project, dryRun); err != nil {
		return err
	}

	// Grant Deployment Manager the required permissions to manage IAM roles.
	if err := setup.SetDMPerms(ctx, project, dryRun); err != nil {
		return err
	}

	// Ensure the DNS zone ends with a period.
	if !strings.HasSuffix(dnsZone, ".") {
		dnsZone = dnsZone + "."
	}

	// Create a deployment with a managed DNS zone and firewall rules which limit SSH to GCE
	// instances to those tunneled over IAP.
	return deployments.Create(ctx, project, "belvedere", setupResources(dnsZone),
		map[string]string{
			"belvedere-type": "base",
		}, dryRun)
}

func setupResources(dnsZone string) []deployments.Resource {
	resources := []deployments.Resource{
		{
			Name: "belvedere-managed-zone",
			Type: "dns.v1.managedZone",
			Properties: dns.ManagedZone{
				Description: fmt.Sprintf("Belvedere managed zone for %s", dnsZone),
				DnsName:     dnsZone,
				Name:        "belvedere",
			},
		},
		{
			Name: "belvedere-deny-ssh-firewall",
			Type: "compute.beta.firewall",
			Properties: compute.Firewall{
				Denied: []*compute.FirewallDenied{
					{
						IPProtocol: "TCP",
						Ports:      []string{"22"},
					},
				},
				Description:  "Deny all SSH to Belvedere apps by default",
				Direction:    "INGRESS",
				Name:         "belvedere-deny-ssh",
				Priority:     65533, // higher than the 65534 of default-allow-ssh
				SourceRanges: []string{"0.0.0.0/0"},
				TargetTags:   []string{"belvedere"},
			},
		},
		{
			Name: "belvedere-iap-tunneling-firewall",
			Type: "compute.beta.firewall",
			Properties: compute.Firewall{
				Allowed: []*compute.FirewallAllowed{
					{
						IPProtocol: "TCP",
						Ports:      []string{"0-65535"},
					},
				},
				Description: "Allow IAP tunneling to Belvedere apps",
				Direction:   "INGRESS",
				Name:        "belvedere-allow-iap",
				Priority:    65532,
				// per https://cloud.google.com/iap/docs/using-tcp-forwarding#starting_ssh
				SourceRanges: []string{"35.235.240.0/20"},
				TargetTags:   []string{"belvedere"},
			},
		},
	}
	return resources
}

func Teardown(ctx context.Context, project string, dryRun, async bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.Teardown")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.BoolAttribute("dry_run", dryRun),
		trace.BoolAttribute("async", async),
	)
	defer span.End()

	return deployments.Delete(ctx, project, "belvedere", dryRun, async)
}
