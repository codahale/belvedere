package belvedere

import (
	"context"
	"strings"
	"testing"

	"github.com/codahale/belvedere/internal/assert"
	"github.com/codahale/belvedere/internal/httpmock"
	"google.golang.org/api/option"
	"google.golang.org/api/secretmanager/v1"
)

func TestSecretsService_List(t *testing.T) {
	t.Parallel()

	srv := httpmock.NewServer(t)
	defer srv.Finish()

	srv.Expect(`/v1/projects/my-project/secrets?alt=json&fields=secrets.name&prettyPrint=false`,
		httpmock.RespJSON(secretmanager.ListSecretsResponse{
			Secrets: []*secretmanager.Secret{
				{
					Name: "one",
				},
				{
					Name: "two",
				},
			},
		}))

	sm, err := secretmanager.NewService(
		context.Background(),
		option.WithEndpoint(srv.URL()),
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

func TestSecretsService_Create(t *testing.T) {
	t.Parallel()

	srv := httpmock.NewServer(t)
	defer srv.Finish()

	srv.Expect(`/v1/projects/my-project/secrets?alt=json&prettyPrint=false&secretId=my-secret`,
		httpmock.ReqJSON(secretmanager.Secret{
			Replication: &secretmanager.Replication{
				Automatic: &secretmanager.Automatic{},
			},
		}),
		httpmock.RespJSON(secretmanager.Secret{}))

	srv.Expect(`/v1/projects/my-project/secrets/my-secret:addVersion?alt=json&prettyPrint=false`,
		httpmock.ReqJSON(
			secretmanager.AddSecretVersionRequest{
				Payload: &secretmanager.SecretPayload{
					Data: "c2VjcmV0",
				},
			}),
		httpmock.RespJSON(secretmanager.SecretVersion{}))

	sm, err := secretmanager.NewService(
		context.Background(),
		option.WithEndpoint(srv.URL()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	secrets := &secretsService{
		project: "my-project",
		sm:      sm,
	}

	if err := secrets.Create(context.Background(), "my-secret", strings.NewReader("secret"), false); err != nil {
		t.Fatal(err)
	}
}

func TestSecretsService_Update(t *testing.T) {
	t.Parallel()

	srv := httpmock.NewServer(t)
	defer srv.Finish()

	srv.Expect(`/v1/projects/my-project/secrets/my-secret:addVersion?alt=json&prettyPrint=false`,
		httpmock.ReqJSON(secretmanager.AddSecretVersionRequest{
			Payload: &secretmanager.SecretPayload{
				Data: "c2VjcmV0",
			},
		}),
		httpmock.RespJSON(secretmanager.SecretVersion{}))

	sm, err := secretmanager.NewService(
		context.Background(),
		option.WithEndpoint(srv.URL()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	secrets := &secretsService{
		project: "my-project",
		sm:      sm,
	}

	if err := secrets.Update(context.Background(), "my-secret", strings.NewReader("secret"), false); err != nil {
		t.Fatal(err)
	}
}

func TestSecretsService_Delete(t *testing.T) {
	t.Parallel()

	srv := httpmock.NewServer(t)
	defer srv.Finish()

	srv.Expect(`/v1/projects/my-project/secrets/my-secret?alt=json&prettyPrint=false`,
		httpmock.Method("delete"),
		httpmock.RespJSON(secretmanager.Empty{}))

	sm, err := secretmanager.NewService(
		context.Background(),
		option.WithEndpoint(srv.URL()),
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

func TestSecretsService_Grant(t *testing.T) {
	t.Parallel()

	srv := httpmock.NewServer(t)
	defer srv.Finish()

	srv.Expect(`/v1/projects/my-project/secrets/my-secret:getIamPolicy?alt=json&prettyPrint=false`,
		httpmock.RespJSON(secretmanager.Policy{
			Etag: "300",
		}))

	srv.Expect(`/v1/projects/my-project/secrets/my-secret:setIamPolicy?alt=json&prettyPrint=false`,
		httpmock.ReqJSON(secretmanager.SetIamPolicyRequest{
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
		}),
		httpmock.RespJSON(secretmanager.Policy{}))

	sm, err := secretmanager.NewService(
		context.Background(),
		option.WithEndpoint(srv.URL()),
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

func TestSecretsService_Revoke(t *testing.T) {
	t.Parallel()

	srv := httpmock.NewServer(t)
	defer srv.Finish()

	srv.Expect(`/v1/projects/my-project/secrets/my-secret:getIamPolicy?alt=json&prettyPrint=false`,
		httpmock.RespJSON(secretmanager.Policy{
			Bindings: []*secretmanager.Binding{
				{
					Role: "roles/secretmanager.secretAccessor",
					Members: []string{
						"serviceAccount:my-app-sa@my-project.iam.gserviceaccount.com",
					},
				},
			},
			Etag: "300",
		}))

	srv.Expect(`/v1/projects/my-project/secrets/my-secret:setIamPolicy?alt=json&prettyPrint=false`,
		httpmock.ReqJSON(secretmanager.SetIamPolicyRequest{
			Policy: &secretmanager.Policy{
				Etag: "300",
			},
		}),
		httpmock.RespJSON(secretmanager.Policy{}))

	sm, err := secretmanager.NewService(
		context.Background(),
		option.WithEndpoint(srv.URL()),
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
