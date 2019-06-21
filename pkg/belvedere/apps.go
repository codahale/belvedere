package belvedere

import (
	"context"
	"fmt"

	"github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
	"go.opencensus.io/trace"
	"google.golang.org/api/deploymentmanager/v2"
	"gopkg.in/yaml.v2"
)

type AppConfig struct {
}

func LoadAppConfig(configPath string) (*AppConfig, error) {
	r, err := openPath(configPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()

	var config AppConfig
	if err := yaml.NewDecoder(r).Decode(&config); err != nil {
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

	resp, err := dm.Deployments.List(project).Filter("labels.type eq belvedere-app").Do()
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

func CreateApp(ctx context.Context, project, appName string, config *AppConfig) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.CreateApp")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", appName),
	)
	defer span.End()

	// TODO create deployment w/ load balancer gubbins

	return errUnimplemented
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
