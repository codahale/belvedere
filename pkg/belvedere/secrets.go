package belvedere

import (
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/internal/gcp"
	"go.opencensus.io/trace"
	"google.golang.org/api/secretmanager/v1"
)

// Secret is a secret stored in Secret Manager.
type Secret struct {
	Name string
}

// SecretsService manages secrets using Google Secret Manager.
type SecretsService interface {
	// List returns a list of all Belvedere-managed secrets.
	List(ctx context.Context) ([]Secret, error)
	// Create creates a new secret with the given name and value.
	Create(ctx context.Context, name string, value []byte, dryRun bool) error
	// Update updates the given secret with a new value.
	Update(ctx context.Context, name string, value []byte, dryRun bool) error
	// Delete deletes the given secret.
	Delete(ctx context.Context, name string, dryRun bool) error
	// Grant modifies the given secret's IAM policy to grant read access to the given app.
	Grant(ctx context.Context, secret, app string, dryRun bool) error
	// Revoke modifies the given secret's IAM policy to revoke read access from the given app.
	Revoke(ctx context.Context, secret, app string, dryRun bool) error
}

type secretsService struct {
	project string
	sm      *secretmanager.Service
}

func (s *secretsService) List(ctx context.Context) ([]Secret, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.secrets.list")
	defer span.End()

	name := fmt.Sprintf("projects/%s", s.project)
	var secrets []Secret
	if err := s.sm.Projects.Secrets.List(name).Fields("secrets.name").Pages(ctx,
		func(list *secretmanager.ListSecretsResponse) error {
			for _, s := range list.Secrets {
				parts := strings.Split(s.Name, "/")
				secrets = append(secrets, Secret{
					Name: parts[len(parts)-1],
				})
			}
			return nil
		},
	); err != nil {
		return nil, fmt.Errorf("error listing secrets: %w", err)
	}
	return secrets, nil
}

func (s *secretsService) Create(ctx context.Context, name string, value []byte, dryRun bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.secrets.create")
	span.AddAttributes(
		trace.StringAttribute("name", name),
		trace.StringAttribute("value", obscure(value)),
	)
	defer span.End()

	if dryRun {
		return nil
	}

	// Create a new version.
	if _, err := s.sm.Projects.Secrets.Create(
		fmt.Sprintf("projects/%s", s.project),
		&secretmanager.Secret{
			Replication: &secretmanager.Replication{Automatic: &secretmanager.Automatic{}},
		},
	).SecretId(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("error creating secret: %w", err)
	}

	// Update the secret's value.
	return s.Update(ctx, name, value, false)
}

func (s *secretsService) Update(ctx context.Context, name string, value []byte, dryRun bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.secrets.update")
	span.AddAttributes(
		trace.StringAttribute("name", name),
		trace.StringAttribute("value", obscure(value)),
	)
	defer span.End()

	if dryRun {
		return nil
	}

	// Add a version to the given secret.
	if _, err := s.sm.Projects.Secrets.AddVersion(
		fmt.Sprintf("projects/%s/secrets/%s", s.project, name),
		&secretmanager.AddSecretVersionRequest{
			Payload: &secretmanager.SecretPayload{Data: base64.StdEncoding.EncodeToString(value)},
		},
	).Context(ctx).Do(); err != nil {
		return fmt.Errorf("error adding secret version: %w", err)
	}
	return nil
}

func (s *secretsService) Delete(ctx context.Context, name string, dryRun bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.secrets.delete")
	span.AddAttributes(
		trace.StringAttribute("name", name),
	)
	defer span.End()

	if dryRun {
		return nil
	}

	_, err := s.sm.Projects.Secrets.
		Delete(fmt.Sprintf("projects/%s/secrets/%s", s.project, name)).
		Context(ctx).Do()
	return err
}

func (s *secretsService) Grant(ctx context.Context, secret, app string, dryRun bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.secrets.grant")
	span.AddAttributes(
		trace.StringAttribute("secret", secret),
		trace.StringAttribute("app", app),
	)
	defer span.End()

	sa := fmt.Sprintf("serviceAccount:%s-sa@%s.iam.gserviceaccount.com", app, s.project)
	return s.modifyIAMPolicy(ctx, fmt.Sprintf("projects/%s/secrets/%s", s.project, secret),
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
		},
		dryRun,
	)
}

func (s *secretsService) Revoke(ctx context.Context, secret, app string, dryRun bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.secrets.revoke")
	span.AddAttributes(
		trace.StringAttribute("app", app),
		trace.StringAttribute("secret", secret),
	)
	defer span.End()

	sa := fmt.Sprintf("serviceAccount:%s-sa@%s.iam.gserviceaccount.com", app, s.project)
	return s.modifyIAMPolicy(ctx, fmt.Sprintf("projects/%s/secrets/%s", s.project, secret),
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
		},
		dryRun,
	)
}

func (s *secretsService) modifyIAMPolicy(
	ctx context.Context,
	secret string,
	f func(policy *secretmanager.Policy) *secretmanager.Policy,
	dryRun bool) error {
	err := gcp.ModifyLoop(5*time.Second, 1*time.Minute, func() error {
		// Get the secret's IAM policy.
		policy, err := s.sm.Projects.Secrets.GetIamPolicy(secret).Context(ctx).Do()
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
		_, err = s.sm.Projects.Secrets.SetIamPolicy(secret, &secretmanager.SetIamPolicyRequest{
			Policy: policy,
		}).Context(ctx).Do()

		return err
	})

	if err != nil {
		return fmt.Errorf("error setting IAM policy: %w", err)
	}
	return nil
}

func obscure(secret []byte) string {
	h := sha512.Sum512(secret)
	return hex.EncodeToString(h[:4])
}

const accessor = "roles/secretmanager.secretAccessor"
