package resources

import (
	"fmt"

	"github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
	compute "google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/dns/v1"
)

func Base(dnsZone string) []deployments.Resource {
	resources := []deployments.Resource{
		// A managed DNS zone for all the app A records.
		{
			Name: "belvedere-managed-zone",
			Type: "dns.v1.managedZone",
			Properties: &dns.ManagedZone{
				Description: fmt.Sprintf("Belvedere managed zone for %s", dnsZone),
				DnsName:     dnsZone,
				Name:        "belvedere",
			},
		},
		// A firewall rule which denies all SSH traffic to belvedere-managed instances.
		{
			Name: "belvedere-deny-ssh",
			Type: "compute.beta.firewall",
			Properties: &compute.Firewall{
				Denied: []*compute.FirewallDenied{
					{
						IPProtocol: "TCP",
						Ports:      []string{"22"},
					},
				},
				Description:  "Deny all SSH to Belvedere apps by default",
				Direction:    "INGRESS",
				Priority:     65533, // higher than the 65534 of default-allow-ssh
				SourceRanges: []string{"0.0.0.0/0"},
				TargetTags:   []string{"belvedere"},
			},
		},
		// A firewall rule which allows all IAP tunnel traffic to belvedere-managed instances.
		{
			Name: "belvedere-allow-iap-tunneling",
			Type: "compute.beta.firewall",
			Properties: &compute.Firewall{
				Allowed: []*compute.FirewallAllowed{
					{
						IPProtocol: "TCP",
						Ports:      []string{"0-65535"},
					},
				},
				Description: "Allow IAP tunneling to Belvedere apps",
				Direction:   "INGRESS",
				Priority:    65532,
				// per https://cloud.google.com/iap/docs/using-tcp-forwarding#starting_ssh
				SourceRanges: []string{"35.235.240.0/20"},
				TargetTags:   []string{"belvedere"},
			},
		},
	}
	return resources
}
