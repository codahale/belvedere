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

	tlsCert, tlsKey, err := cert.GenerateSelfSignedCertKey(fmt.Sprintf("belvedere-%s.blort"), nil, nil)
	if err != nil {
		return err
	}

	config := &deployments.Config{
		Resources: []deployments.Resource{
			// A firewall rule allowing access from the load balancer to app instances on port 8443.
			{
				Name: "firewall",
				Type: "compute.beta.firewall",
				Properties: compute.Firewall{
					Name: fmt.Sprintf("belvedere-app-%s", appName),
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
				Name: "healthcheck",
				Type: "compute.beta.healthCheck",
				Properties: compute.HealthCheck{
					Name: fmt.Sprintf("belvedere-%s", appName),
					Type: "HTTP2",
					Http2HealthCheck: &compute.HTTP2HealthCheck{
						PortName:    "svc-https",
						RequestPath: "/healthz",
					},
				},
			},
			// An HTTP2 backend service connected to the healthcheck.
			{
				Name: "backend-service",
				Type: "compute.beta.backendService",
				Properties: compute.BackendService{
					Name:     fmt.Sprintf("belvedere-%s", appName),
					Protocol: "HTTP2",
					PortName: "svc-https",
					HealthChecks: []string{
						deployments.SelfLink("healthcheck"),
					},
					Region: region,
				},
			},
			// A URL map directing requests to the backend service while blocking access to the
			// app's health check URL.
			{
				Name: "url-map",
				Type: "compute.beta.urlMap",
				Properties: compute.UrlMap{
					Name:           fmt.Sprintf("belvedere-%s", appName),
					DefaultService: deployments.SelfLink("backend-service"),
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
				Name: "tls-cert",
				Type: "compute.beta.sslCertificate",
				Properties: compute.SslCertificate{
					Name:        fmt.Sprintf("belvedere-%s", appName),
					Certificate: string(tlsCert),
					PrivateKey:  string(tlsKey),
				},
			},
			// A QUIC-enabled HTTPS target proxy using the app's TLS cert and directing requests to
			// the URL map.
			{
				Name: "target-proxy",
				Type: "compute.beta.targetHttpsProxy",
				Properties: compute.TargetHttpsProxy{
					Name: fmt.Sprintf("belvedere-%s", appName),
					SslCertificates: []string{
						deployments.SelfLink("tls-cert"),
					},
					QuicOverride: "ENABLE",
					UrlMap:       deployments.SelfLink("url-map"),
				},
			},
			// A global forwarding rule directing TCP:443 to the target proxy.
			{
				Name: "global-forwarding-rule",
				Type: "compute.beta.globalForwardingRule",
				Properties: compute.ForwardingRule{
					Name:       fmt.Sprintf("belvedere-%s", appName),
					IPProtocol: "TCP",
					PortRange:  "443",
					Target:     deployments.SelfLink("target-proxy"),
				},
			},
			// A service account.
			{
				Name: "service-account",
				Type: "iam.v1.serviceAccount",
				Properties: map[string]string{
					"accountId":   fmt.Sprintf("app-%s", appName),
					"displayName": appName,
				},
			},
		},
	}

	config.Resources = append(config.Resources, roleBindings(project, appName, app)...)

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
