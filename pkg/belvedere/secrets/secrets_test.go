package secrets

import (
	"context"
	"net/http"
	"testing"

	"github.com/codahale/belvedere/pkg/belvedere/internal/it"
	"github.com/google/go-cmp/cmp"
	secretmanager "google.golang.org/api/secretmanager/v1beta1"
	"gopkg.in/h2non/gock.v1"
)

func TestList(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://secretmanager.googleapis.com/v1beta1/projects/my-project/secrets?alt=json&fields=secrets.name&prettyPrint=false").
		Reply(http.StatusOK).
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

	secrets, err := NewService(context.TODO(), "my-project")
	if err != nil {
		t.Fatal(err)
	}

	actual, err := secrets.List(context.TODO())
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

func TestCreate(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://secretmanager.googleapis.com/v1beta1/projects/my-project/secrets?alt=json&prettyPrint=false&secretId=my-secret").
		JSON(secretmanager.Secret{
			Replication: &secretmanager.Replication{
				Automatic: &secretmanager.Automatic{},
			},
		}).
		Reply(http.StatusOK).
		JSON(secretmanager.Secret{})

	gock.New("https://secretmanager.googleapis.com/v1beta1/projects/my-project/secrets/my-secret:addVersion?alt=json&prettyPrint=false").
		JSON(secretmanager.AddSecretVersionRequest{
			Payload: &secretmanager.SecretPayload{
				Data: "c2VjcmV0",
			},
		}).
		Reply(http.StatusOK).
		JSON(secretmanager.SecretVersion{})

	secrets, err := NewService(context.TODO(), "my-project")
	if err != nil {
		t.Fatal(err)
	}

	if err := secrets.Create(context.TODO(), "my-secret", []byte("secret"), false); err != nil {
		t.Fatal(err)
	}
}

func TestUpdate(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://secretmanager.googleapis.com/v1beta1/projects/my-project/secrets/my-secret:addVersion?alt=json&prettyPrint=false").
		JSON(secretmanager.AddSecretVersionRequest{
			Payload: &secretmanager.SecretPayload{
				Data: "c2VjcmV0",
			},
		}).
		Reply(http.StatusOK).
		JSON(secretmanager.SecretVersion{})

	secrets, err := NewService(context.TODO(), "my-project")
	if err != nil {
		t.Fatal(err)
	}

	if err := secrets.Update(context.TODO(), "my-secret", []byte("secret"), false); err != nil {
		t.Fatal(err)
	}
}

func TestDelete(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://secretmanager.googleapis.com/v1beta1/projects/my-project/secrets/my-secret?alt=json&prettyPrint=false").
		Delete("").
		Reply(http.StatusOK).
		JSON(secretmanager.Empty{})

	secrets, err := NewService(context.TODO(), "my-project")
	if err != nil {
		t.Fatal(err)
	}

	if err := secrets.Delete(context.TODO(), "my-secret", false); err != nil {
		t.Fatal(err)
	}
}

func TestGrant(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://secretmanager.googleapis.com/v1beta1/projects/my-project/secrets/my-secret:getIamPolicy?alt=json&prettyPrint=false").
		Reply(http.StatusOK).
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
		Reply(http.StatusOK).
		JSON(secretmanager.Policy{})

	secrets, err := NewService(context.TODO(), "my-project")
	if err != nil {
		t.Fatal(err)
	}

	if err := secrets.Grant(context.TODO(), "my-secret", "my-app", false); err != nil {
		t.Fatal(err)
	}
}

func TestRevoke(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://secretmanager.googleapis.com/v1beta1/projects/my-project/secrets/my-secret:getIamPolicy?alt=json&prettyPrint=false").
		Reply(http.StatusOK).
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
		Reply(http.StatusOK).
		JSON(secretmanager.Policy{})

	secrets, err := NewService(context.TODO(), "my-project")
	if err != nil {
		t.Fatal(err)
	}

	if err := secrets.Revoke(context.TODO(), "my-secret", "my-app", false); err != nil {
		t.Fatal(err)
	}
}
