package setup

import (
	"context"
	"fmt"

	"github.com/codahale/belvedere/pkg/belvedere/internal/gcp"
	"go.opencensus.io/trace"
	"google.golang.org/api/cloudresourcemanager/v1"
)

// SetDMPerms binds the Deployment Manager service account to the `owner` role if it has not already
// been so bound. This allows Deployment Manager to add IAM roles to service accounts per
// https://cloud.google.com/deployment-manager/docs/configuration/set-access-control-resources#granting_deployment_manager_permission_to_set_iam_policies
func SetDMPerms(ctx context.Context, project string, dryRun bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.internal.setup.SetDMPerms")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.BoolAttribute("dry_run", dryRun),
	)
	defer span.End()

	// Get our CRM client.
	crm, err := gcp.CloudResourceManager(ctx)
	if err != nil {
		return err
	}

	// Resolve the project's numeric ID.
	p, err := crm.Projects.Get(project).Fields("projectNumber").Context(ctx).Do()
	if err != nil {
		return err
	}

	// Get the project's IAM policy.
	policy, err := crm.Projects.GetIamPolicy(project, &cloudresourcemanager.GetIamPolicyRequest{}).
		Context(ctx).Do()
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

	// Early exit if we don't want side effects.
	if dryRun {
		return nil
	}

	// Set the modified policy.
	_, err = crm.Projects.SetIamPolicy(project, &cloudresourcemanager.SetIamPolicyRequest{
		Policy: policy,
	}).Context(ctx).Do()
	return err
}
