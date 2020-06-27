package setup

import (
	"context"
	"fmt"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/internal/gcp"
	"go.opencensus.io/trace"
	"google.golang.org/api/cloudresourcemanager/v1"
)

func (s *service) SetDMPerms(ctx context.Context, project string, dryRun bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.internal.setup.SetDMPerms")
	defer span.End()

	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.BoolAttribute("dry_run", dryRun),
	)

	// Resolve the project's numeric ID.
	p, err := s.crm.Projects.Get(project).Fields("projectNumber").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("error getting project: %w", err)
	}

	span.AddAttributes(trace.Int64Attribute("project_number", p.ProjectNumber))

	member := fmt.Sprintf("serviceAccount:%d@cloudservices.gserviceaccount.com", p.ProjectNumber)

	return modifyIAMPolicy(ctx, s.crm, project,
		func(policy *cloudresourcemanager.Policy) *cloudresourcemanager.Policy {
			// Look for an existing IAM binding giving Deployment Service ownership of the project.
			exists := bindingExists(policy.Bindings, member)
			span.AddAttributes(trace.BoolAttribute("binding_exists", exists))

			// If it exists, early exit.
			if exists {
				return nil
			}

			// If none exists, add a binding and update the policy.
			policy.Bindings = append(policy.Bindings, &cloudresourcemanager.Binding{
				Members: []string{member},
				Role:    owner,
			})
			return policy
		},
	)
}

func bindingExists(bindings []*cloudresourcemanager.Binding, member string) bool {
	for _, binding := range bindings {
		if binding.Role == owner {
			for _, m := range binding.Members {
				if m == member {
					return true
				}
			}
		}
	}

	return false
}

const owner = "roles/owner"

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
