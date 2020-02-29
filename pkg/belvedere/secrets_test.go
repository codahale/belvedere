package belvedere

import (
	"context"
	"net/http"
	"testing"

	"github.com/codahale/belvedere/pkg/belvedere/internal/it"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/secretmanager/v1"
	"gopkg.in/h2non/gock.v1"
)

func TestSecretsService_List(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://secretmanager.googleapis.com/v1/projects/my-project/secrets?alt=json&fields=secrets.name&prettyPrint=false").
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

	sm, err := secretmanager.NewService(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	secrets := &secretsService{
		project: "my-project",
		sm:      sm,
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

func TestSecretsService_Create(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://secretmanager.googleapis.com/v1/projects/my-project/secrets?alt=json&prettyPrint=false&secretId=my-secret").
		JSON(secretmanager.Secret{
			Replication: &secretmanager.Replication{
				Automatic: &secretmanager.Automatic{},
			},
		}).
		Reply(http.StatusOK).
		JSON(secretmanager.Secret{})

	gock.New("https://secretmanager.googleapis.com/v1/projects/my-project/secrets/my-secret:addVersion?alt=json&prettyPrint=false").
		JSON(secretmanager.AddSecretVersionRequest{
			Payload: &secretmanager.SecretPayload{
				Data: "c2VjcmV0",
			},
		}).
		Reply(http.StatusOK).
		JSON(secretmanager.SecretVersion{})

	sm, err := secretmanager.NewService(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	secrets := &secretsService{
		project: "my-project",
		sm:      sm,
	}

	if err := secrets.Create(context.TODO(), "my-secret", []byte("secret"), false); err != nil {
		t.Fatal(err)
	}
}

func TestSecretsService_Update(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://secretmanager.googleapis.com/v1/projects/my-project/secrets/my-secret:addVersion?alt=json&prettyPrint=false").
		JSON(secretmanager.AddSecretVersionRequest{
			Payload: &secretmanager.SecretPayload{
				Data: "c2VjcmV0",
			},
		}).
		Reply(http.StatusOK).
		JSON(secretmanager.SecretVersion{})

	sm, err := secretmanager.NewService(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	secrets := &secretsService{
		project: "my-project",
		sm:      sm,
	}

	if err := secrets.Update(context.TODO(), "my-secret", []byte("secret"), false); err != nil {
		t.Fatal(err)
	}
}

func TestSecretsService_Delete(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://secretmanager.googleapis.com/v1/projects/my-project/secrets/my-secret?alt=json&prettyPrint=false").
		Delete("").
		Reply(http.StatusOK).
		JSON(secretmanager.Empty{})

	sm, err := secretmanager.NewService(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	secrets := &secretsService{
		project: "my-project",
		sm:      sm,
	}

	if err := secrets.Delete(context.TODO(), "my-secret", false); err != nil {
		t.Fatal(err)
	}
}

func TestSecretsService_Grant(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://secretmanager.googleapis.com/v1/projects/my-project/secrets/my-secret:getIamPolicy?alt=json&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(secretmanager.Policy{
			Etag: "300",
		})

	gock.New("https://secretmanager.googleapis.com/v1/projects/my-project/secrets/my-secret:setIamPolicy?alt=json&prettyPrint=false").
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

	sm, err := secretmanager.NewService(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	secrets := &secretsService{
		project: "my-project",
		sm:      sm,
	}

	if err := secrets.Grant(context.TODO(), "my-secret", "my-app", false); err != nil {
		t.Fatal(err)
	}
}

func TestSecretsService_Revoke(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://secretmanager.googleapis.com/v1/projects/my-project/secrets/my-secret:getIamPolicy?alt=json&prettyPrint=false").
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

	gock.New("https://secretmanager.googleapis.com/v1/projects/my-project/secrets/my-secret:setIamPolicy?alt=json&prettyPrint=false").
		JSON(
			secretmanager.SetIamPolicyRequest{
				Policy: &secretmanager.Policy{
					Etag: "300",
				},
			}).
		Reply(http.StatusOK).
		JSON(secretmanager.Policy{})

	sm, err := secretmanager.NewService(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	secrets := &secretsService{
		project: "my-project",
		sm:      sm,
	}

	if err := secrets.Revoke(context.TODO(), "my-secret", "my-app", false); err != nil {
		t.Fatal(err)
	}
}
