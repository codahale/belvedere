package resources

import (
	"fmt"
	"strings"

	"github.com/codahale/belvedere/pkg/belvedere/cfg"
	"github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
	"google.golang.org/api/dns/v1"
)

func Name(s ...string) string {
	if len(s) == 0 {
		return "belvedere"
	}

	return fmt.Sprintf("belvedere-%s", strings.Join(s, "-"))
}

type Builder interface {
	// Base returns a list of resources for the base deployment.
	Base(dnsZone string) []deployments.Resource

	// App returns a list of resources for an app deployment.
	App(project string, app string, managedZone *dns.ManagedZone, config *cfg.Config) []deployments.Resource

	// Release returns a list of resources for a release deployment.
	Release(project, region, app, release, imageSHA256 string, config *cfg.Config) []deployments.Resource
}

func NewBuilder() Builder {
	return &builder{}
}

type builder struct {
}
