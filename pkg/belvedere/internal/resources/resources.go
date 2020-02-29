package resources

import (
	"fmt"
	"strings"

	"github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
	compute "google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/dns/v1"
)

func Name(s ...string) string {
	return fmt.Sprintf("belvedere-%s", strings.Join(s, "-"))
}

type Builder interface {
	// Base returns a list of resources for the base deployment.
	Base(dnsZone string) []deployments.Resource

	// App returns a list of resources for an app deployment.
	App(
		project string, app string, managedZone *dns.ManagedZone,
		cdn *compute.BackendServiceCdnPolicy, iap *compute.BackendServiceIAP,
		iamRoles []string,
	) []deployments.Resource

	// Release returns a list of resources for a release deployment.
	Release(
		project, region, app, release, network, subnetwork, machineType, userData string,
		replicas int, autoscalingPolicy *compute.AutoscalingPolicy,
	) []deployments.Resource
}

func NewBuilder() Builder {
	return &builder{}
}

type builder struct {
}
