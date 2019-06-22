package setup

import (
	"context"
	"fmt"

	"go.opencensus.io/trace"
	"google.golang.org/api/cloudresourcemanager/v1"
)

// SetDMPerms binds the Deployment Manager service account to the `owner` role if it has not already
// been so bound. This allows Deployment Manager to add IAM roles to service accounts per
// https://cloud.google.com/deployment-manager/docs/configuration/set-access-control-resources#granting_deployment_manager_permission_to_set_iam_policies
func SetDMPerms(ctx context.Context, project string) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.internal.setup.SetDMPerms")
	defer span.End()

	crm, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return err
	}

	// Resolve the project's numeric ID.
	p, err := crm.Projects.Get(project).Fields("projectNumber").Do()
	if err != nil {
		return err
	}

	// Get the project's IAM policy.
	policy, err := crm.Projects.GetIamPolicy(project, &cloudresourcemanager.GetIamPolicyRequest{}).Do()
	if err != nil {
		return err
	}

	crmMember := fmt.Sprintf("serviceAccount:%d@cloudservices.gserviceaccount.com", p.ProjectNumber)
	const owner = "roles/owner"

	// Look for an existing IAM binding giving Deployment Manager ownership of the project.
	for _, binding := range policy.Bindings {
		if binding.Role == owner {
			for _, member := range binding.Members {
				if member == crmMember {
					span.Annotate(
						[]trace.Attribute{
							trace.Int64Attribute("project_number", p.ProjectNumber),
						},
						"Binding verified",
					)
					return nil
				}
			}
		}
	}

	// If none exists, add a binding and update the policy.
	span.Annotate(
		[]trace.Attribute{
			trace.Int64Attribute("project_number", p.ProjectNumber),
		},
		"Binding created",
	)
	policy.Bindings = append(policy.Bindings, &cloudresourcemanager.Binding{
		Members: []string{crmMember},
		Role:    owner,
	})
	_, err = crm.Projects.SetIamPolicy(project, &cloudresourcemanager.SetIamPolicyRequest{
		Policy: policy,
	}).Do()
	return err
}
