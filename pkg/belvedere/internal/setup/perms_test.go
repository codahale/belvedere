package setup

import (
	"context"
	"net/http"
	"testing"

	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/option"
	"gopkg.in/h2non/gock.v1"
)

//nolint:paralleltest // uses Gock
func TestManager_SetDMPerms(t *testing.T) {
	defer gock.Off()

	gock.New("https://cloudresourcemanager.googleapis.com/v1/projects/" +
		"my-project?alt=json&fields=projectNumber&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(cloudresourcemanager.Project{
			ProjectNumber: 123456,
		})

	gock.New("https://cloudresourcemanager.googleapis.com/v1/projects/" +
		"my-project:getIamPolicy?alt=json&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(cloudresourcemanager.Policy{
			Bindings: []*cloudresourcemanager.Binding{
				{
					Members: []string{"email:existing@example.com"},
					Role:    "roles/passerby",
				},
			},
			Etag: "300",
		})

	gock.New("https://cloudresourcemanager.googleapis.com/v1/projects/" +
		"my-project:getIamPolicy?alt=json&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(cloudresourcemanager.Policy{
			Bindings: []*cloudresourcemanager.Binding{
				{
					Members: []string{"email:existing@example.com"},
					Role:    "roles/passerby",
				},
			},
			Etag: "301",
		})

	gock.New("https://cloudresourcemanager.googleapis.com/v1/projects/" +
		"my-project:setIamPolicy?alt=json&prettyPrint=false").
		JSON(cloudresourcemanager.SetIamPolicyRequest{
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
		}).
		Reply(http.StatusConflict)

	gock.New("https://cloudresourcemanager.googleapis.com/v1/projects/" +
		"my-project:setIamPolicy?alt=json&prettyPrint=false").
		JSON(cloudresourcemanager.SetIamPolicyRequest{
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
		}).
		Reply(http.StatusOK).
		JSON(cloudresourcemanager.Policy{})

	crm, err := cloudresourcemanager.NewService(
		context.Background(),
		option.WithHTTPClient(http.DefaultClient),
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

//nolint:paralleltest // uses Gock
func TestManager_SetDMPermsExisting(t *testing.T) {
	defer gock.Off()

	gock.New("https://cloudresourcemanager.googleapis.com/v1/projects/" +
		"my-project?alt=json&fields=projectNumber&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(cloudresourcemanager.Project{
			ProjectNumber: 123456,
		})

	gock.New("https://cloudresourcemanager.googleapis.com/v1/projects/" +
		"my-project:getIamPolicy?alt=json&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(cloudresourcemanager.Policy{
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
		})

	crm, err := cloudresourcemanager.NewService(
		context.Background(),
		option.WithHTTPClient(http.DefaultClient),
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
