package belvedere

import (
	"context"
	"errors"
	"fmt"

	"github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
	"go.opencensus.io/trace"
	"google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/deploymentmanager/v2"
	"google.golang.org/api/dns/v1"
)

func ListApps(ctx context.Context, project string) ([]string, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.ListApps")
	span.AddAttributes(
		trace.StringAttribute("project", project),
	)
	defer span.End()

	dm, err := deploymentmanager.NewService(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := dm.Deployments.List(project).Do()
	if err != nil {
		return nil, err
	}

	var names []string
	for _, d := range resp.Deployments {
		var app bool
		var name string
		for _, l := range d.Labels {
			if l.Key == "belvedere-app" {
				name = l.Value
			} else if l.Key == "belvedere-type" && l.Value == "app" {
				app = true
			}
		}
		if app {
			names = append(names, name)
		}
	}
	return names, nil
}

func CreateApp(ctx context.Context, project, region, appName string, config *Config, dryRun bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.CreateApp")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("region", region),
		trace.StringAttribute("app", appName),
		trace.BoolAttribute("dry_run", dryRun),
	)
	defer span.End()

	managedZone, err := findManagedZone(ctx, project)
	if err != nil {
		return err
	}

	resources := appResources(project, appName, managedZone, config)

	name := fmt.Sprintf("belvedere-%s", appName)
	return deployments.Insert(ctx, project, name, resources, map[string]string{
		"belvedere-type":   "app",
		"belvedere-app":    appName,
		"belvedere-region": region,
	}, dryRun)
}

func UpdateApp(ctx context.Context, project, appName string, config *Config, dryRun bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.CreateApp")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", appName),
		trace.BoolAttribute("dry_run", dryRun),
	)
	defer span.End()

	managedZone, err := findManagedZone(ctx, project)
	if err != nil {
		return err
	}

	resources := appResources(project, appName, managedZone, config)

	name := fmt.Sprintf("belvedere-%s", appName)
	return deployments.Update(ctx, project, name, resources, dryRun)
}

func DestroyApp(ctx context.Context, project, appName string, dryRun bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.DestroyApp")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", appName),
	)
	defer span.End()

	return deployments.Delete(ctx, project, fmt.Sprintf("belvedere-%s", appName), dryRun)
}

func findManagedZone(ctx context.Context, project string) (*dns.ManagedZone, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.findManagedZone")
	span.AddAttributes(
		trace.StringAttribute("project", project),
	)
	defer span.End()

	d, err := dns.NewService(ctx)
	if err != nil {
		return nil, err
	}

	return d.ManagedZones.Get(project, "belvedere").Do()
}

func findRegion(ctx context.Context, project, appName string) (string, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.findRegion")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", appName),
	)
	defer span.End()

	dm, err := deploymentmanager.NewService(ctx)
	if err != nil {
		return "", err
	}

	deployment, err := dm.Deployments.Get(project, fmt.Sprintf("belvedere-%s", appName)).Do()
	if err != nil {
		return "", err
	}

	for _, l := range deployment.Labels {
		if l.Key == "belvedere-region" {
			return l.Value, nil
		}
	}

	return "", errors.New("no region found")
}

func appResources(project string, appName string, managedZone *dns.ManagedZone, config *Config) *deployments.Config {
	firewall := fmt.Sprintf("%s-fw", appName)
	healthcheck := fmt.Sprintf("%s-hc", appName)
	backendService := fmt.Sprintf("%s-bes", appName)
	urlMap := fmt.Sprintf("%s-urlmap", appName)
	sslCertificate := fmt.Sprintf("%s-cert", appName)
	targetProxy := fmt.Sprintf("%s-tp", appName)
	forwardingRule := fmt.Sprintf("%s-fr", appName)
	serviceAccount := fmt.Sprintf("%s-sa", appName)
	dnsRecord := fmt.Sprintf("%s-rrs", appName)
	dnsName := fmt.Sprintf("%s.%s", appName, managedZone.DnsName)
	dmConfig := &deployments.Config{
		Resources: []deployments.Resource{
			// A firewall rule allowing access from the load balancer to app instances on port 8443.
			{
				Name: firewall,
				Type: "compute.beta.firewall",
				Properties: compute.Firewall{
					Allowed: []*compute.FirewallAllowed{
						{
							IPProtocol: "TCP",
							Ports:      []string{"8443"},
						},
					},
					SourceRanges: []string{
						"130.211.0.0/22",
						"35.191.0.0/16",
					},
					TargetTags: []string{
						fmt.Sprintf("belvedere-%s", appName),
					},
				},
			},
			// An HTTP2 healthcheck which sends a request to the svc-https named port for the path
			// /healthz.
			{
				Name: healthcheck,
				Type: "compute.beta.healthCheck",
				Properties: compute.HealthCheck{
					Type: "HTTP2",
					Http2HealthCheck: &compute.HTTP2HealthCheck{
						PortName:    "svc-https",
						RequestPath: "/healthz",
					},
				},
			},
			// An HTTP2 backend service connected to the healthcheck.
			{
				Name: backendService,
				Type: "compute.beta.backendService",
				Properties: compute.BackendService{
					Protocol: "HTTP2",
					PortName: "svc-https",
					HealthChecks: []string{
						deployments.SelfLink(healthcheck),
					},
				},
			},
			// A URL map directing requests to the backend service while blocking access to the
			// app's health check URL.
			{
				Name: urlMap,
				Type: "compute.beta.urlMap",
				Properties: compute.UrlMap{
					DefaultService: deployments.SelfLink(backendService),
					// TODO add WAF rule turning /healthz from external sources into 404
					//PathMatchers: []*compute.PathMatcher{
					//	{
					//		Name: "deny-external-healthchecks",
					//		PathRules: []*compute.PathRule{
					//			{
					//				Paths:   []string{"/healthz"},
					//				Service: deployments.SelfLink("backend-service"),
					//				RouteAction: &compute.HttpRouteAction{
					//					FaultInjectionPolicy: &compute.HttpFaultInjection{
					//						Abort: &compute.HttpFaultAbort{
					//							HttpStatus: 404,
					//						},
					//					},
					//				},
					//			},
					//		},
					//	},
					//},
				},
			},
			// A TLS certificate.
			{
				Name: sslCertificate,
				Type: "compute.beta.sslCertificate",
				Properties: compute.SslCertificate{
					Managed: &compute.SslCertificateManagedSslCertificate{
						Domains: []string{dnsName},
					},
					Type: "MANAGED",
				},
			},
			// A QUIC-enabled HTTPS target proxy using the app's TLS cert and directing requests to
			// the URL map.
			{
				Name: targetProxy,
				Type: "compute.beta.targetHttpsProxy",
				Properties: compute.TargetHttpsProxy{
					SslCertificates: []string{
						deployments.SelfLink(sslCertificate),
					},
					QuicOverride: "ENABLE",
					UrlMap:       deployments.SelfLink(urlMap),
				},
			},
			// A global forwarding rule directing TCP:443 to the target proxy.
			{
				Name: forwardingRule,
				Type: "compute.beta.globalForwardingRule",
				Properties: compute.ForwardingRule{
					IPProtocol: "TCP",
					PortRange:  "443",
					Target:     deployments.SelfLink(targetProxy),
				},
			},
			// A service account.
			{
				Name: serviceAccount,
				Type: "iam.v1.serviceAccount",
				Properties: map[string]string{
					"accountId":   fmt.Sprintf("app-%s", appName),
					"displayName": appName,
				},
			},
			// A DNS record.
			{
				Name: dnsRecord,
				Type: "gcp-types/dns-v1:resourceRecordSets",
				Properties: map[string]interface{}{
					"name":        dnsName,
					"managedZone": managedZone.Name,
					"records": []*dns.ResourceRecordSet{
						{
							Type:    "A",
							Rrdatas: []string{deployments.Ref(forwardingRule, "IPAddress")},
							Ttl:     50,
						},
					},
				},
			},
		},
	}
	dmConfig.Resources = append(dmConfig.Resources, roleBindings(project, serviceAccount, config.IAMRoles)...)
	return dmConfig
}
