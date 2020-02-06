package belvedere

import (
	"context"
	"testing"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/internal/it"
	"github.com/codahale/belvedere/pkg/belvedere/internal/waiter"
	"github.com/google/go-cmp/cmp"
	secretmanager "google.golang.org/api/secretmanager/v1beta1"
	"gopkg.in/h2non/gock.v1"
)

func TestSecrets(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://secretmanager.googleapis.com/v1beta1/projects/my-project/secrets?alt=json&prettyPrint=false").
		Reply(200).
		JSON(secretmanager.ListSecretsResponse{
			Secrets: []*secretmanager.Secret{
				{
					Name: "one",
					Replication: &secretmanager.Replication{
						Automatic: &secretmanager.Automatic{},
					},
				},
				{
					Name: "two",
					Replication: &secretmanager.Replication{
						UserManaged: &secretmanager.UserManaged{
							Replicas: []*secretmanager.Replica{
								{
									Location: "us-east1",
								},
								{
									Location: "us-west1",
								},
							},
						},
					},
				},
			},
		})

	ctx := waiter.WithInterval(context.TODO(), 10*time.Millisecond)
	actual, err := Secrets(ctx, "my-project")
	if err != nil {
		t.Fatal(err)
	}

	expected := []Secret{
		{
			Name:        "one",
			Replication: "automatic",
		},
		{
			Name:        "two",
			Replication: "user-managed: [us-east1 us-west1]",
		},
	}

	if !cmp.Equal(expected, actual) {
		t.Fatal(cmp.Diff(expected, actual))
	}
}

func TestGrantSecret(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://secretmanager.googleapis.com/v1beta1/projects/my-project/secrets/my-secret:getIamPolicy?alt=json&prettyPrint=false").
		Reply(200).
		JSON(secretmanager.Policy{
			Etag: "300",
		})

	gock.New("https://secretmanager.googleapis.com/v1beta1/projects/my-project/secrets/my-secret:setIamPolicy?alt=json&prettyPrint=false").
		JSON(
			secretmanager.SetIamPolicyRequest{
				Policy: &secretmanager.Policy{
					Bindings: []*secretmanager.Binding{
						{
							Role: "roles/secretmanager.secretAccessor",
							Members: []string{
								"serviceAccount:my-app-sa@my-project.iam.gserviceaccount.com",
							},
						},
					},
					Etag: "300",
				},
			}).
		Reply(200).
		JSON(secretmanager.Policy{})

	ctx := waiter.WithInterval(context.TODO(), 10*time.Millisecond)
	if err := GrantSecret(ctx, "my-project", "my-app", "my-secret", false); err != nil {
		t.Fatal(err)
	}
}

func TestRevokeSecret(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://secretmanager.googleapis.com/v1beta1/projects/my-project/secrets/my-secret:getIamPolicy?alt=json&prettyPrint=false").
		Reply(200).
		JSON(secretmanager.Policy{
			Bindings: []*secretmanager.Binding{
				{
					Role: "roles/secretmanager.secretAccessor",
					Members: []string{
						"serviceAccount:my-app-sa@my-project.iam.gserviceaccount.com",
					},
				},
			},
			Etag: "300",
		})

	gock.New("https://secretmanager.googleapis.com/v1beta1/projects/my-project/secrets/my-secret:setIamPolicy?alt=json&prettyPrint=false").
		JSON(
			secretmanager.SetIamPolicyRequest{
				Policy: &secretmanager.Policy{
					Etag: "300",
				},
			}).
		Reply(200).
		JSON(secretmanager.Policy{})

	ctx := waiter.WithInterval(context.TODO(), 10*time.Millisecond)
	if err := RevokeSecret(ctx, "my-project", "my-app", "my-secret", false); err != nil {
		t.Fatal(err)
	}
}
