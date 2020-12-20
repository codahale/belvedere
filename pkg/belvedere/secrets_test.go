package belvedere

import (
	"context"
	"net/http"
	"testing"

	"github.com/codahale/belvedere/internal/assert"
	"google.golang.org/api/option"
	"google.golang.org/api/secretmanager/v1"
	"gopkg.in/h2non/gock.v1"
)

//nolint:paralleltest // uses Gock
func TestSecretsService_List(t *testing.T) {
	defer gock.Off()

	gock.New("https://secretmanager.googleapis.com/v1/projects/my-project/" +
		"secrets?alt=json&fields=secrets.name&prettyPrint=false").
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

	sm, err := secretmanager.NewService(
		context.Background(),
		option.WithHTTPClient(http.DefaultClient),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	secrets := &secretsService{
		project: "my-project",
		sm:      sm,
	}

	got, err := secrets.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	want := []Secret{
		{
			Name: "one",
		},
		{
			Name: "two",
		},
	}

	assert.Equal(t, "List()", want, got)
}

//nolint:paralleltest // uses Gock
func TestSecretsService_Create(t *testing.T) {
	defer gock.Off()

	gock.New("https://secretmanager.googleapis.com/v1/projects/my-project/" +
		"secrets?alt=json&prettyPrint=false&secretId=my-secret").
		JSON(secretmanager.Secret{
			Replication: &secretmanager.Replication{
				Automatic: &secretmanager.Automatic{},
			},
		}).
		Reply(http.StatusOK).
		JSON(secretmanager.Secret{})

	gock.New("https://secretmanager.googleapis.com/v1/projects/my-project/" +
		"secrets/my-secret:addVersion?alt=json&prettyPrint=false").
		JSON(secretmanager.AddSecretVersionRequest{
			Payload: &secretmanager.SecretPayload{
				Data: "c2VjcmV0",
			},
		}).
		Reply(http.StatusOK).
		JSON(secretmanager.SecretVersion{})

	sm, err := secretmanager.NewService(
		context.Background(),
		option.WithHTTPClient(http.DefaultClient),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	secrets := &secretsService{
		project: "my-project",
		sm:      sm,
	}

	if err := secrets.Create(context.Background(), "my-secret", []byte("secret"), false); err != nil {
		t.Fatal(err)
	}
}

//nolint:paralleltest // uses Gock
func TestSecretsService_Update(t *testing.T) {
	defer gock.Off()

	gock.New("https://secretmanager.googleapis.com/v1/projects/my-project/secrets/" +
		"my-secret:addVersion?alt=json&prettyPrint=false").
		JSON(secretmanager.AddSecretVersionRequest{
			Payload: &secretmanager.SecretPayload{
				Data: "c2VjcmV0",
			},
		}).
		Reply(http.StatusOK).
		JSON(secretmanager.SecretVersion{})

	sm, err := secretmanager.NewService(
		context.Background(),
		option.WithHTTPClient(http.DefaultClient),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	secrets := &secretsService{
		project: "my-project",
		sm:      sm,
	}

	if err := secrets.Update(context.Background(), "my-secret", []byte("secret"), false); err != nil {
		t.Fatal(err)
	}
}

//nolint:paralleltest // uses Gock
func TestSecretsService_Delete(t *testing.T) {
	defer gock.Off()

	gock.New("https://secretmanager.googleapis.com/v1/projects/my-project/secrets/" +
		"my-secret?alt=json&prettyPrint=false").
		Delete("").
		Reply(http.StatusOK).
		JSON(secretmanager.Empty{})

	sm, err := secretmanager.NewService(
		context.Background(),
		option.WithHTTPClient(http.DefaultClient),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	secrets := &secretsService{
		project: "my-project",
		sm:      sm,
	}

	if err := secrets.Delete(context.Background(), "my-secret", false); err != nil {
		t.Fatal(err)
	}
}

//nolint:paralleltest // uses Gock
func TestSecretsService_Grant(t *testing.T) {
	defer gock.Off()

	gock.New("https://secretmanager.googleapis.com/v1/projects/my-project/secrets/" +
		"my-secret:getIamPolicy?alt=json&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(secretmanager.Policy{
			Etag: "300",
		})

	gock.New("https://secretmanager.googleapis.com/v1/projects/my-project/secrets/" +
		"my-secret:setIamPolicy?alt=json&prettyPrint=false").
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

	sm, err := secretmanager.NewService(
		context.Background(),
		option.WithHTTPClient(http.DefaultClient),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	secrets := &secretsService{
		project: "my-project",
		sm:      sm,
	}

	if err := secrets.Grant(context.Background(), "my-secret", "my-app", false); err != nil {
		t.Fatal(err)
	}
}

//nolint:paralleltest // uses Gock
func TestSecretsService_Revoke(t *testing.T) {
	defer gock.Off()

	gock.New("https://secretmanager.googleapis.com/v1/projects/my-project/secrets/" +
		"my-secret:getIamPolicy?alt=json&prettyPrint=false").
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

	gock.New("https://secretmanager.googleapis.com/v1/projects/my-project/secrets/" +
		"my-secret:setIamPolicy?alt=json&prettyPrint=false").
		JSON(
			secretmanager.SetIamPolicyRequest{
				Policy: &secretmanager.Policy{
					Etag: "300",
				},
			}).
		Reply(http.StatusOK).
		JSON(secretmanager.Policy{})

	sm, err := secretmanager.NewService(
		context.Background(),
		option.WithHTTPClient(http.DefaultClient),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	secrets := &secretsService{
		project: "my-project",
		sm:      sm,
	}

	if err := secrets.Revoke(context.Background(), "my-secret", "my-app", false); err != nil {
		t.Fatal(err)
	}
}
