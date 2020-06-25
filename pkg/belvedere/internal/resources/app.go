package resources

import (
	"fmt"
	"math"

	"github.com/codahale/belvedere/pkg/belvedere/cfg"
	"github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/dns/v1"
)

//nolint:funlen // not worth splitting this up
func (*builder) App(
	project string, app string, managedZone *dns.ManagedZone, config *cfg.Config,
) []deployments.Resource {
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
	securityPolicy := fmt.Sprintf("%s-waf", app)

	resources := []deployments.Resource{
		// A firewall rule allowing access from the load balancer to application instances on port
		// 8443.
		{
			Name: firewall,
			Type: "compute.v1.firewall",
			Properties: &compute.Firewall{
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
					Name(app),
				},
			},
		},
		// An HTTP2 healthcheck which sends a request to the svc-https named port for the path
		// /healthz.
		{
			Name: healthcheck,
			Type: "compute.v1.healthCheck",
			Properties: &compute.HealthCheck{
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
			Type: "compute.v1.backendService",
			Properties: &compute.BackendService{
				CdnPolicy: config.CDNPolicy,
				ConnectionDraining: &compute.ConnectionDraining{
					DrainingTimeoutSec: 60,
				},
				EnableCDN: config.CDNPolicy != nil,
				HealthChecks: []string{
					deployments.SelfLink(healthcheck),
				},
				Iap: config.IAP,
				LogConfig: &compute.BackendServiceLogConfig{
					Enable: true,
				},
				PortName:        "svc-https",
				Protocol:        "HTTP2",
				SecurityPolicy:  deployments.SelfLink(securityPolicy),
				SessionAffinity: sessionAffinityEnum(config.SessionAffinity),
			},
		},
		// A URL map directing requests to the backend service while blocking access to the
		// app's health check URL.
		{
			Name: urlMap,
			Type: "compute.v1.urlMap",
			Properties: &compute.UrlMap{
				DefaultService: deployments.SelfLink(backendService),
			},
		},
		// A TLS certificate.
		{
			Name: sslCertificate,
			Type: "compute.v1.sslCertificate",
			Properties: &compute.SslCertificate{
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
			Type: "compute.v1.targetHttpsProxy",
			Properties: &compute.TargetHttpsProxy{
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
			Type: "compute.v1.globalForwardingRule",
			Properties: &compute.ForwardingRule{
				IPProtocol: "TCP",
				PortRange:  "443",
				Target:     deployments.SelfLink(targetProxy),
			},
		},
		// A service account.
		{
			Name: serviceAccount,
			Type: "iam.v1.serviceAccount",
			Properties: &deployments.ServiceAccount{
				AccountID:   fmt.Sprintf("app-%s", app),
				DisplayName: app,
			},
		},
		// A DNS record.
		{
			Name: dnsRecord,
			Type: "gcp-types/dns-v1:resourceRecordSets",
			Properties: &deployments.ResourceRecordSets{
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
		// The Cloud WAF security policy.
		{
			Name: securityPolicy,
			Type: "compute.v1.securityPolicy",
			Properties: &compute.SecurityPolicy{
				Description: fmt.Sprintf("WAF rules for Belvedere app %s.", app),
				Rules: append(config.WAFRules,
					&compute.SecurityPolicyRule{
						Action:      "deny(404)",
						Description: "Deny external access to healthchecks.",
						Match: &compute.SecurityPolicyRuleMatcher{
							Expr: &compute.Expr{
								Expression: "request.path.matches('^/healthz')",
							},
						},
						Priority: 1,
					},
					&compute.SecurityPolicyRule{
						Action:      "allow",
						Description: "Allow all access by default.",
						Match: &compute.SecurityPolicyRuleMatcher{
							Config: &compute.SecurityPolicyRuleMatcherConfig{
								SrcIpRanges: []string{"*"},
							},
							VersionedExpr: "SRC_IPS_V1",
						},
						Priority: math.MaxInt32,
					},
				),
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

// requiresRoles is a list of IAM role which are added to application service accounts by default.
//nolint:gochecknoglobals // can't have non-scalar consts
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

// roleBinding returns a Deployment Manager IAM member binding for the given service account and
// role.
func roleBinding(project, serviceAccount, role string) deployments.Resource {
	return deployments.Resource{
		Name: fmt.Sprintf("%s-%s", serviceAccount, role),
		Type: "gcp-types/cloudresourcemanager-v1:virtual.projects.iamMemberBinding",
		Properties: &deployments.IAMMemberBinding{
			Resource: project,
			Role:     role,
			Member:   fmt.Sprintf("serviceAccount:%s", deployments.Ref(serviceAccount, "email")),
		},
	}
}

func sessionAffinityEnum(s string) string {
	switch s {
	case cfg.SessionAffinityCookie:
		return "GENERATED_COOKIE"
	case cfg.SessionAffinityIP:
		return "CLIENT_IP"
	default:
		return "NONE"
	}
}
