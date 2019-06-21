package belvedere

import (
	"context"
	"fmt"

	"github.com/codahale/belvedere/pkg/belvedere/internal/base"
	"github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
	"go.opencensus.io/trace"
	compute "google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/dns/v1"
)

func Initialize(ctx context.Context, projectID, dnsZone string) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.cli.Initialize")
	span.AddAttributes(
		trace.StringAttribute("project", projectID),
		trace.StringAttribute("dns_zone", dnsZone),
	)
	defer span.End()

	// Enable all required services.
	if err := base.EnableServices(ctx, projectID); err != nil {
		return err
	}

	// Grant Deployment Manager the required permissions to manage IAM roles.
	if err := base.SetDMPerms(ctx, projectID); err != nil {
		return err
	}

	// Create a deployment with a managed DNS zone and firewall rules which limit SSH to GCE
	// instances to those tunneled over IAP.
	config := &deployments.Config{
		Resources: []deployments.Resource{
			{
				Name: "managed-zone",
				Type: "dns.v1.managedZone",
				Properties: dns.ManagedZone{
					Description: fmt.Sprintf("Belvedere managed zone for %s", dnsZone),
					DnsName:     dnsZone,
					Name:        "belvedere",
				},
			},
			{
				Name: "deny-ssh-firewall",
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
				Name: "iap-tunneling-firewall",
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
		},
	}

	return deployments.Create(ctx, projectID, "belvedere", config, map[string]string{"belvedere-type": "base"})
}
