package setup

import (
	"context"
	"net/http"
	"testing"

	"github.com/codahale/belvedere/pkg/belvedere/internal/it"
	"google.golang.org/api/cloudresourcemanager/v1"
	"gopkg.in/h2non/gock.v1"
)

func TestSetDMPerms(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://cloudresourcemanager.googleapis.com/v1/projects/my-project?alt=json&fields=projectNumber&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(cloudresourcemanager.Project{
			ProjectNumber: 123456,
		})

	gock.New("https://cloudresourcemanager.googleapis.com/v1/projects/my-project:getIamPolicy?alt=json&prettyPrint=false").
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

	gock.New("https://cloudresourcemanager.googleapis.com/v1/projects/my-project:getIamPolicy?alt=json&prettyPrint=false").
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

	gock.New("https://cloudresourcemanager.googleapis.com/v1/projects/my-project:setIamPolicy?alt=json&prettyPrint=false").
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

	gock.New("https://cloudresourcemanager.googleapis.com/v1/projects/my-project:setIamPolicy?alt=json&prettyPrint=false").
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

	if err := SetDMPerms(context.TODO(), "my-project", false); err != nil {
		t.Fatal(err)
	}
}
