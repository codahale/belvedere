package belvedere

import (
	"context"

	"go.opencensus.io/trace"
	"google.golang.org/api/deploymentmanager/v2"
)

type Env struct {
	Name    string
	DNSName string
}

func CreateEnv(ctx context.Context, projectID string, envName, dnsName string) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.CreateEnv")
	span.AddAttributes(
		trace.StringAttribute("project_id", projectID),
		trace.StringAttribute("env.name", envName),
		trace.StringAttribute("env.dns_name", dnsName),
	)
	defer span.End()

	return createDeployment(ctx, deployment{
		name:      envName,
		projectID: projectID,
		config: config{
			Resources: []resource{
				managedZone(envName, dnsName),
			},
		},
		labels: map[string]string{
			"type": "belvedere-env",
			"name": envName,
		},
	})
}

func ListEnvs(ctx context.Context, projectID string) ([]string, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.ListEnvs")
	span.AddAttributes(trace.StringAttribute("project_id", projectID))
	defer span.End()

	dm, err := deploymentmanager.NewService(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := dm.Deployments.List(projectID).Filter(`labels.type eq belvedere-env`).Do()
	if err != nil {
		return nil, err
	}

	envs := make([]string, len(resp.Deployments))
	for i, d := range resp.Deployments {
		envs[i] = d.Name
	}
	return envs, nil
}

func DestroyEnv(ctx context.Context, projectID string, envName string) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.DestroyEnv")
	span.AddAttributes(trace.StringAttribute("project_id", projectID))
	defer span.End()

	return deleteDeployment(ctx, projectID, envName)
}
