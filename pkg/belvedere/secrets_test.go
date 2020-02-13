package belvedere

import (
	"context"
	"testing"

	"github.com/codahale/belvedere/pkg/belvedere/internal/it"
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
				},
				{
					Name: "two",
				},
			},
		})

	actual, err := Secrets(context.TODO(), "my-project")
	if err != nil {
		t.Fatal(err)
	}

	expected := []Secret{
		{
			Name: "one",
		},
		{
			Name: "two",
		},
	}

	if !cmp.Equal(expected, actual) {
		t.Fatal(cmp.Diff(expected, actual))
	}
}

func TestCreateSecret(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://secretmanager.googleapis.com/v1beta1/projects/my-project/secrets?alt=json&prettyPrint=false&secretId=my-secret").
		JSON(secretmanager.Secret{
			Replication: &secretmanager.Replication{
				Automatic: &secretmanager.Automatic{},
			},
		}).
		Reply(200).
		JSON(secretmanager.Secret{})

	gock.New("https://secretmanager.googleapis.com/v1beta1/projects/my-project/secrets/my-secret:addVersion?alt=json&prettyPrint=false").
		JSON(secretmanager.AddSecretVersionRequest{
			Payload: &secretmanager.SecretPayload{
				Data: "c2VjcmV0",
			},
		}).
		Reply(200).
		JSON(secretmanager.SecretVersion{})

	if err := CreateSecret(context.TODO(), "my-project", "my-secret", []byte("secret")); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateSecret(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://secretmanager.googleapis.com/v1beta1/projects/my-project/secrets/my-secret:addVersion?alt=json&prettyPrint=false").
		JSON(secretmanager.AddSecretVersionRequest{
			Payload: &secretmanager.SecretPayload{
				Data: "c2VjcmV0",
			},
		}).
		Reply(200).
		JSON(secretmanager.SecretVersion{})

	if err := UpdateSecret(context.TODO(), "my-project", "my-secret", []byte("secret")); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteSecret(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://secretmanager.googleapis.com/v1beta1/projects/my-project/secrets/my-secret?alt=json&prettyPrint=false").
		Delete("").
		Reply(200).
		JSON(secretmanager.Empty{})

	if err := DeleteSecret(context.TODO(), "my-project", "my-secret"); err != nil {
		t.Fatal(err)
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

	if err := GrantSecret(context.TODO(), "my-project", "my-secret", "my-app", false); err != nil {
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

	if err := RevokeSecret(context.TODO(), "my-project", "my-app", "my-secret", false); err != nil {
		t.Fatal(err)
	}
}
