package belvedere

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/codahale/belvedere/pkg/belvedere/internal/gcp"
	"go.opencensus.io/trace"
	"google.golang.org/api/googleapi"
	secretmanager "google.golang.org/api/secretmanager/v1beta1"
)

const accessor = "roles/secretmanager.secretAccessor"

func GrantAppSecret(ctx context.Context, project, app, secret string, dryRun bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.GrantAppSecret")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", app),
		trace.StringAttribute("secret", secret),
		trace.BoolAttribute("dry_run", dryRun),
	)
	defer span.End()

	sm, err := gcp.SecretManager(ctx)
	if err != nil {
		return err
	}

	sa := fmt.Sprintf("serviceAccount:%s-sa@%s.iam.gserviceaccount.com", app, project)
	return modifyIAMPolicy(ctx, sm, fmt.Sprintf("projects/%s/secrets/%s", project, secret), dryRun,
		func(policy *secretmanager.Policy) *secretmanager.Policy {
			// Look for an existing IAM binding giving the app access to the secret.
			for _, binding := range policy.Bindings {
				if binding.Role == accessor {
					for _, member := range binding.Members {
						if member == sa {
							span.Annotate(
								[]trace.Attribute{
									trace.StringAttribute("service_account", sa),
								},
								"Binding verified",
							)
							return nil
						}
					}
				}
			}

			// If none exists, add a binding and update the policy.
			policy.Bindings = append(policy.Bindings, &secretmanager.Binding{
				Members: []string{sa},
				Role:    accessor,
			})
			span.Annotate(
				[]trace.Attribute{
					trace.StringAttribute("service_account", sa),
				},
				"Binding created",
			)
			return policy

		})
}

func RevokeAppSecret(ctx context.Context, project, app, secret string, dryRun bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.RevokeAppSecret")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", app),
		trace.StringAttribute("secret", secret),
		trace.BoolAttribute("dry_run", dryRun),
	)
	defer span.End()

	sm, err := gcp.SecretManager(ctx)
	if err != nil {
		return err
	}

	sa := fmt.Sprintf("serviceAccount:%s-sa@%s.iam.gserviceaccount.com", app, project)
	return modifyIAMPolicy(ctx, sm, fmt.Sprintf("projects/%s/secrets/%s", project, secret), dryRun,
		func(policy *secretmanager.Policy) *secretmanager.Policy {
			var bindings []*secretmanager.Binding

			// Copy everything that's not an IAM binding giving the app access to the secret.
			for _, binding := range policy.Bindings {
				if binding.Role == accessor {
					remove := false
					for _, member := range binding.Members {
						if member == sa {
							span.Annotate(
								[]trace.Attribute{
									trace.StringAttribute("service_account", sa),
								},
								"Binding removed",
							)
							remove = true
							break
						}
					}

					if !remove {
						bindings = append(bindings, binding)
					}
				}
			}

			// If no such policy exists, nevermind.
			if len(bindings) == len(policy.Bindings) {
				span.Annotate(
					[]trace.Attribute{
						trace.StringAttribute("service_account", sa),
					},
					"Binding not found",
				)
				return nil
			}

			// Otherwise, update the policy.
			policy.Bindings = bindings
			return policy
		})
}

func modifyIAMPolicy(
	ctx context.Context,
	sm *secretmanager.Service,
	secret string,
	dryRun bool,
	f func(policy *secretmanager.Policy) *secretmanager.Policy) error {

	bo := backoff.NewExponentialBackOff()
	bo.MaxInterval = 5 * time.Second
	bo.MaxElapsedTime = 1 * time.Minute
	for {
		// Get the secret's IAM policy.
		policy, err := sm.Projects.Secrets.GetIamPolicy(secret).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("error getting IAM policy: %w", err)
		}

		// Modify the secret's IAM policy.
		policy = f(policy)
		if policy == nil {
			return nil
		}

		// Don't modify anything if it's a dry run.
		if dryRun {
			return nil
		}

		// Set the modified policy.
		_, err = sm.Projects.Secrets.SetIamPolicy(secret, &secretmanager.SetIamPolicyRequest{
			Policy: policy,
		}).Context(ctx).Do()

		// If the policy was modified underneath us, try again.
		if e, ok := err.(*googleapi.Error); ok {
			if e.Code == 409 {
				d := bo.NextBackOff()
				if d == backoff.Stop {
					return fmt.Errorf("couldn't write IAM policy after %s", bo.GetElapsedTime())
				}
				time.Sleep(d)
				continue
			}
		} else if err != nil {
			return fmt.Errorf("error setting IAM policy: %w", err)
		}
		return nil
	}
}
