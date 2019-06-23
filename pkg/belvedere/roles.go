package belvedere

import (
	"fmt"

	"github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
)

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

func roleBindings(project, serviceAccount string, roles []string) []deployments.Resource {
	var bindings []deployments.Resource
	for _, role := range requiredRoles {
		bindings = append(bindings, roleBinding(project, serviceAccount, role))
	}
	for _, role := range roles {
		bindings = append(bindings, roleBinding(project, serviceAccount, role))
	}
	return bindings
}
