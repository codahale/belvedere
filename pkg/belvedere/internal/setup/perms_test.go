package setup

import (
	"context"
	"net/http"
	"testing"

	"github.com/codahale/gubbins/httpmock"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/option"
)

func TestManager_SetDMPerms(t *testing.T) {
	t.Parallel()

	srv := httpmock.NewServer(t)
	defer srv.Finish()

	srv.Expect(`/v1/projects/my-project?alt=json&fields=projectNumber&prettyPrint=false`,
		httpmock.RespJSON(cloudresourcemanager.Project{
			ProjectNumber: 123456,
		}))

	srv.Expect(`/v1/projects/my-project:getIamPolicy?alt=json&prettyPrint=false`,
		httpmock.RespJSON(cloudresourcemanager.Policy{
			Bindings: []*cloudresourcemanager.Binding{
				{
					Members: []string{"email:existing@example.com"},
					Role:    "roles/passerby",
				},
			},
			Etag: "300",
		}))

	srv.Expect(`/v1/projects/my-project:getIamPolicy?alt=json&prettyPrint=false`,
		httpmock.RespJSON(cloudresourcemanager.Policy{
			Bindings: []*cloudresourcemanager.Binding{
				{
					Members: []string{"email:existing@example.com"},
					Role:    "roles/passerby",
				},
			},
			Etag: "301",
		}))

	srv.Expect(`/v1/projects/my-project:setIamPolicy?alt=json&prettyPrint=false`,
		httpmock.ReqJSON(cloudresourcemanager.SetIamPolicyRequest{
			Policy: &cloudresourcemanager.Policy{
				Bindings: []*cloudresourcemanager.Binding{
					{
						Members: []string{"email:existing@example.com"},
						Role:    "roles/passerby",
					},
					{
						Members: []string{"serviceAccount:123456@cloudservices.gserviceaccount.com"},
						Role:    "roles/owner",
					},
				},
				Etag: "300",
			},
		}),
		httpmock.Status(http.StatusConflict))

	srv.Expect(`/v1/projects/my-project:setIamPolicy?alt=json&prettyPrint=false`,
		httpmock.ReqJSON(cloudresourcemanager.SetIamPolicyRequest{
			Policy: &cloudresourcemanager.Policy{
				Bindings: []*cloudresourcemanager.Binding{
					{
						Members: []string{"email:existing@example.com"},
						Role:    "roles/passerby",
					},
					{
						Members: []string{"serviceAccount:123456@cloudservices.gserviceaccount.com"},
						Role:    "roles/owner",
					},
				},
				Etag: "301",
			},
		}),
		httpmock.RespJSON(cloudresourcemanager.Policy{}))

	crm, err := cloudresourcemanager.NewService(
		context.Background(),
		option.WithEndpoint(srv.URL()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	m := &service{crm: crm}

	if err := m.SetDMPerms(context.Background(), "my-project", false); err != nil {
		t.Fatal(err)
	}
}

func TestManager_SetDMPermsExisting(t *testing.T) {
	t.Parallel()

	srv := httpmock.NewServer(t)
	defer srv.Finish()

	srv.Expect(`/v1/projects/my-project?alt=json&fields=projectNumber&prettyPrint=false`,
		httpmock.RespJSON(cloudresourcemanager.Project{
			ProjectNumber: 123456,
		}))

	srv.Expect(`/v1/projects/my-project:getIamPolicy?alt=json&prettyPrint=false`,
		httpmock.RespJSON(cloudresourcemanager.Policy{
			Bindings: []*cloudresourcemanager.Binding{
				{
					Members: []string{"email:existing@example.com"},
					Role:    "roles/passerby",
				},
				{
					Members: []string{"serviceAccount:123456@cloudservices.gserviceaccount.com"},
					Role:    "roles/owner",
				},
			},
			Etag: "300",
		}))

	crm, err := cloudresourcemanager.NewService(
		context.Background(),
		option.WithEndpoint(srv.URL()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	m := &service{crm: crm}

	if err := m.SetDMPerms(context.Background(), "my-project", false); err != nil {
		t.Fatal(err)
	}
}
