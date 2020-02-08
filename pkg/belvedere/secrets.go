package belvedere

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/codahale/belvedere/pkg/belvedere/internal/gcp"
	"go.opencensus.io/trace"
	"google.golang.org/api/googleapi"
	secretmanager "google.golang.org/api/secretmanager/v1beta1"
)

// Secret is a secret stored in Secret Manager.
type Secret struct {
	Name string
}

// Secrets returns a list of all secrets stored in Secret Manager for the given project.
func Secrets(ctx context.Context, project string) ([]Secret, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.Secrets")
	span.AddAttributes(
		trace.StringAttribute("project", project),
	)
	defer span.End()

	sm, err := gcp.SecretManager(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := sm.Projects.Secrets.List(fmt.Sprintf("projects/%s", project)).
		Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("error listing secrets: %w", err)
	}

	var secrets []Secret
	for _, s := range resp.Secrets {
		parts := strings.Split(s.Name, "/")
		secrets = append(secrets, Secret{
			Name: parts[len(parts)-1],
		})
	}
	return secrets, nil
}

// CreateSecret creates a secret with the given name and value. If the path is a filename, the
// contents of the file are used as the value. If the path is `-`, the contents of STDIN are used.
func CreateSecret(ctx context.Context, project, secret, path string) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.CreateSecret")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("secret", secret),
		trace.StringAttribute("path", path),
	)
	defer span.End()

	// Create a Secret Manager client.
	sm, err := gcp.SecretManager(ctx)
	if err != nil {
		return err
	}

	// Create a new version.
	_, err = sm.Projects.Secrets.Create(fmt.Sprintf("projects/%s", project),
		&secretmanager.Secret{
			Replication: &secretmanager.Replication{Automatic: &secretmanager.Automatic{}},
		}).SecretId(secret).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("error creating secret: %w", err)
	}

	// Update the secret's value.
	return UpdateSecret(ctx, project, secret, path)
}

// UpdateSecret updates the secret's value. If the path is a filename, the contents of the file are
// used as the value. If the path is `-`, the contents of STDIN are used.
func UpdateSecret(ctx context.Context, project, secret, path string) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.UpdateSecret")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("secret", secret),
		trace.StringAttribute("path", path),
	)
	defer span.End()

	// Either open the file or use STDIN.
	var r io.ReadCloser
	if path == "-" {
		r = os.Stdin
	} else {
		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("error opening %s: %w", path, err)
		}
		r = f
	}
	defer func() { _ = r.Close() }()

	// Read the entire input.
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", path, err)
	}

	// Create a Secret Manager client.
	sm, err := gcp.SecretManager(ctx)
	if err != nil {
		return err
	}

	// Add a version to the given secret.
	_, err = sm.Projects.Secrets.AddVersion(fmt.Sprintf("projects/%s/secrets/%s", project, secret),
		&secretmanager.AddSecretVersionRequest{
			Payload: &secretmanager.SecretPayload{Data: base64.StdEncoding.EncodeToString(b)},
		}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("error adding secret version: %w", err)
	}
	return nil
}

// DeleteSecret deletes the given secret.
func DeleteSecret(ctx context.Context, project, secret string) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.CreateSecret")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("secret", secret),
	)
	defer span.End()

	sm, err := gcp.SecretManager(ctx)
	if err != nil {
		return err
	}

	_, err = sm.Projects.Secrets.Delete(fmt.Sprintf("projects/%s/secrets/%s", project, secret)).
		Context(ctx).Do()
	return err
}

const accessor = "roles/secretmanager.secretAccessor"

// GrantSecret modifies the IAM policy of the given secret to allow the given application's
// service account access.
func GrantSecret(ctx context.Context, project, secret, app string, dryRun bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.GrantSecret")
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

// RevokeSecret modifies the IAM policy of the given secret to deny the given application's
// service account access to it.
func RevokeSecret(ctx context.Context, project, app, secret string, dryRun bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.RevokeSecret")
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
