package belvedere

import (
	"context"
	"errors"
	"fmt"

	"github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
	"github.com/codahale/belvedere/pkg/belvedere/internal/gcp"
	"go.opencensus.io/trace"
	"google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/dns/v1"
)

type App struct {
	Project string
	Region  string
	Name    string
}

// ListApps returns a list of apps which have been created in the given project.
func ListApps(ctx context.Context, project string) ([]App, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.ListApps")
	span.AddAttributes(
		trace.StringAttribute("project", project),
	)
	defer span.End()

	list, err := deployments.List(ctx, project)
	if err != nil {
		return nil, err
	}

	var apps []App
	for _, labels := range list {
		if labels["belvedere-type"] == "app" {
			apps = append(apps, App{
				Project: project,
				Region:  labels["belvedere-region"],
				Name:    labels["belvedere-app"],
			})
		}
	}
	return apps, nil
}

// CreateApp creates an app in the given project and region with the given name and configuration.
func CreateApp(ctx context.Context, project, region, app string, config *Config, dryRun bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.CreateApp")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("region", region),
		trace.StringAttribute("app", app),
		trace.BoolAttribute("dry_run", dryRun),
	)
	defer span.End()

	if err := validateRFC1035(app); err != nil {
		return err
	}

	managedZone, err := findManagedZone(ctx, project)
	if err != nil {
		return err
	}

	resources := appResources(project, app, managedZone, config)

	name := fmt.Sprintf("belvedere-%s", app)
	return deployments.Create(ctx, project, name, resources, map[string]string{
		"belvedere-type":   "app",
		"belvedere-app":    app,
		"belvedere-region": region,
	}, dryRun)
}

// UpdateApp updates the resources for the given app to match the given configuration.
func UpdateApp(ctx context.Context, project, app string, config *Config, dryRun bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.CreateApp")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", app),
		trace.BoolAttribute("dry_run", dryRun),
	)
	defer span.End()

	managedZone, err := findManagedZone(ctx, project)
	if err != nil {
		return err
	}

	resources := appResources(project, app, managedZone, config)

	name := fmt.Sprintf("belvedere-%s", app)
	return deployments.Update(ctx, project, name, resources, dryRun)
}

// DeleteApp deletes all the resources associated with the given app.
func DeleteApp(ctx context.Context, project, app string, dryRun, async bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.DeleteApp")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", app),
		trace.BoolAttribute("dry_run", dryRun),
		trace.BoolAttribute("async", async),
	)
	defer span.End()

	return deployments.Delete(ctx, project, fmt.Sprintf("belvedere-%s", app), dryRun, async)
}

func findManagedZone(ctx context.Context, project string) (*dns.ManagedZone, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.findManagedZone")
	span.AddAttributes(
		trace.StringAttribute("project", project),
	)
	defer span.End()

	d, err := gcp.DNS(ctx)
	if err != nil {
		return nil, err
	}

	return d.ManagedZones.Get(project, "belvedere").Do()
}

func findRegion(ctx context.Context, project, app string) (string, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.findRegion")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", app),
	)
	defer span.End()

	dm, err := gcp.DeploymentManager(ctx)
	if err != nil {
		return "", err
	}

	deployment, err := dm.Deployments.Get(project, fmt.Sprintf("belvedere-%s", app)).Do()
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

func appResources(project string, app string, managedZone *dns.ManagedZone, config *Config) []deployments.Resource {
	firewall := fmt.Sprintf("belvedere-allow-%s-lb", app)
	healthcheck := fmt.Sprintf("%s-hc", app)
	backendService := fmt.Sprintf("%s-bes", app)
	urlMap := fmt.Sprintf("%s-urlmap", app)
	sslCertificate := fmt.Sprintf("%s-cert", app)
	targetProxy := fmt.Sprintf("%s-tp", app)
	forwardingRule := fmt.Sprintf("%s-fr", app)
	serviceAccount := fmt.Sprintf("%s-sa", app)
	dnsRecord := fmt.Sprintf("%s-rrs", app)
	dnsName := fmt.Sprintf("%s.%s", app, managedZone.DnsName)
	resources := []deployments.Resource{
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
					fmt.Sprintf("belvedere-%s", app),
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
			Properties: deployments.ServiceAccount{
				AccountID:   fmt.Sprintf("app-%s", app),
				DisplayName: app,
			},
		},
		// A DNS record.
		{
			Name: dnsRecord,
			Type: "gcp-types/dns-v1:resourceRecordSets",
			Properties: deployments.ResourceRecordSets{
				Name:        dnsName,
				ManagedZone: managedZone.Name,
				Records: []*dns.ResourceRecordSet{
					{
						Type:    "A",
						Rrdatas: []string{deployments.Ref(forwardingRule, "IPAddress")},
						Ttl:     50,
					},
				},
			},
		},
	}

	for _, role := range requiredRoles {
		resources = append(resources, roleBinding(project, serviceAccount, role))
	}

	for _, role := range config.IAMRoles {
		resources = append(resources, roleBinding(project, serviceAccount, role))
	}

	return resources
}

var requiredRoles = []string{
	"roles/clouddebugger.agent",
	"roles/cloudprofiler.agent",
	"roles/cloudtrace.agent",
	"roles/errorreporting.writer",
	"roles/logging.logWriter",
	"roles/monitoring.metricWriter",
	"roles/stackdriver.resourceMetadata.writer",
	"roles/storage.objectViewer",
}

func roleBinding(project, serviceAccount, role string) deployments.Resource {
	return deployments.Resource{
		Name: fmt.Sprintf("%s-%s", serviceAccount, role),
		Type: "gcp-types/cloudresourcemanager-v1:virtual.projects.iamMemberBinding",
		Properties: map[string]string{
			"resource": project,
			"role":     role,
			"member":   fmt.Sprintf("serviceAccount:%s", deployments.Ref(serviceAccount, "email")),
		},
	}
}
