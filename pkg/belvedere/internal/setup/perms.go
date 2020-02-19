package setup

import (
	"context"
	"fmt"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/internal/gcp"
	"go.opencensus.io/trace"
	"google.golang.org/api/cloudresourcemanager/v1"
)

func modifyIAMPolicy(
	ctx context.Context,
	crm *cloudresourcemanager.Service,
	project string,
	f func(policy *cloudresourcemanager.Policy) *cloudresourcemanager.Policy) error {
	err := gcp.ModifyLoop(5*time.Second, 1*time.Minute, func() error {
		// Get the project's IAM policy.
		policy, err := crm.Projects.GetIamPolicy(project, &cloudresourcemanager.GetIamPolicyRequest{}).
			Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("error getting IAM policy: %w", err)
		}

		// Modify the project's IAM policy.
		policy = f(policy)
		if policy == nil {
			return nil
		}

		// Set the modified policy.
		_, err = crm.Projects.SetIamPolicy(project, &cloudresourcemanager.SetIamPolicyRequest{
			Policy: policy,
		}).Context(ctx).Do()
		return err
	})

	if err != nil {
		return fmt.Errorf("error modifying IAM policy: %w", err)
	}
	return nil
}

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
		return fmt.Errorf("error getting project: %w", err)
	}

	member := fmt.Sprintf("serviceAccount:%d@cloudservices.gserviceaccount.com", p.ProjectNumber)
	const owner = "roles/owner"

	exists := false
	err = modifyIAMPolicy(ctx, crm, project, func(policy *cloudresourcemanager.Policy) *cloudresourcemanager.Policy {
		// Look for an existing IAM binding giving Deployment Manager ownership of the project.
		for _, binding := range policy.Bindings {
			if binding.Role == owner {
				for _, m := range binding.Members {
					if m == member {
						exists = true
						return nil
					}
				}
			}
		}

		// If none exists, add a binding and update the policy.
		policy.Bindings = append(policy.Bindings, &cloudresourcemanager.Binding{
			Members: []string{member},
			Role:    owner,
		})
		return policy
	})
	if err != nil {
		return err
	}

	msg := "Binding created"
	if exists {
		msg = "Binding verified"
	}
	span.Annotate(
		[]trace.Attribute{
			trace.Int64Attribute("project_number", p.ProjectNumber),
		},
		msg,
	)

	return nil
}
