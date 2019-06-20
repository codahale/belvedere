package belvedere

import (
	"google.golang.org/api/dns/v1"
)

func managedZone(envName, dnsName string) resource {
	return resource{
		Name: "managed-zone",
		Type: "dns.v1.managedZone",
		Properties: dns.ManagedZone{
			Description: envName,
			DnsName:     dnsName,
			Labels: map[string]string{
				"env": envName,
			},
			Name: envName,
		},
	}
}
