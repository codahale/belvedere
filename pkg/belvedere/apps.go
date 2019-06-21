package belvedere

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
	"go.opencensus.io/trace"
	"google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/deploymentmanager/v2"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/util/cert"
)

type AppConfig struct {
	IAMRoles []string `yaml:"iam_roles,omitempty"`
}

func LoadAppConfig(ctx context.Context, path string) (*AppConfig, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.LoadAppConfig")
	span.AddAttributes(
		trace.StringAttribute("path", path),
	)
	defer span.End()

	r, err := openPath(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var config AppConfig
	if err := yaml.Unmarshal(b, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

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

func CreateApp(ctx context.Context, project, region, appName string, app *AppConfig) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.CreateApp")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", appName),
	)
	defer span.End()

	certPEM, keyPEM, err := cert.GenerateSelfSignedCertKey(fmt.Sprintf("belvedere-%s.blort"), nil, nil)
	if err != nil {
		return err
	}

	firewall := fmt.Sprintf("%s-fw", appName)
	healthcheck := fmt.Sprintf("%s-hc", appName)
	backendService := fmt.Sprintf("%s-bes", appName)
	urlMap := fmt.Sprintf("%s-urlmap", appName)
	sslCertificate := fmt.Sprintf("%s-cert", appName)
	targetProxy := fmt.Sprintf("%s-tp", appName)
	forwardingRule := fmt.Sprintf("%s-fr", appName)
	serviceAccount := fmt.Sprintf("%s-sa", appName)

	config := &deployments.Config{
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
						"0.0.0.0/0", // TODO lock down access to apps
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
					Region: region,
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
				// TODO replace self-signed certs with managed certs
				// https://cloud.google.com/load-balancing/docs/ssl-certificates#google-managed_ssl_certificate_renewal
				Name: sslCertificate,
				Type: "compute.beta.sslCertificate",
				Properties: compute.SslCertificate{
					Certificate: string(certPEM),
					PrivateKey:  string(keyPEM),
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
		},
	}

	config.Resources = append(config.Resources, roleBindings(project, serviceAccount, app)...)

	name := fmt.Sprintf("belvedere-%s", appName)
	return deployments.Insert(ctx, project, name, config, map[string]string{
		"belvedere-type": "app",
		"belvedere-app":  appName,
	})
}

func DestroyApp(ctx context.Context, project, appName string) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.DestroyApp")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", appName),
	)
	defer span.End()

	return deployments.Delete(ctx, project, fmt.Sprintf("belvedere-%s", appName))
}
