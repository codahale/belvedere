package belvedere

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
	"go.opencensus.io/trace"
	compute "google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/deploymentmanager/v2"
	"gopkg.in/yaml.v2"
)

type AppConfig struct {
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

	resp, err := dm.Deployments.List(project).Filter("labels.belvedere-type eq app").Do()
	if err != nil {
		return nil, err
	}

	var names []string
	for _, d := range resp.Deployments {
		for _, l := range d.Labels {
			if l.Key == "belvedere-name" {
				names = append(names, l.Value)
			}
		}
	}
	return names, nil
}

func CreateApp(ctx context.Context, project, appName string, app *AppConfig) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.CreateApp")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", appName),
	)
	defer span.End()

	config := &deployments.Config{
		Resources: []deployments.Resource{
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
					TargetTags: []string{
						fmt.Sprintf("belvedere-app-%s", appName),
					},
				},
			},
			// TODO add global forwarding rule
			// TODO add target proxy
			// TODO add url map
			// TODO add backend service
			// TODO add health check
			// TODO add managed SSL cert
			// TODO add service account w/ role bindings
		},
	}

	name := fmt.Sprintf("belvedere-%s", appName)
	return deployments.Insert(ctx, project, name, config, map[string]string{
		"belvedere-type": "app",
		"belvedere-name": appName,
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
