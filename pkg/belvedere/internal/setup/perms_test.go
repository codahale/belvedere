package setup

import (
	"context"
	"testing"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/internal/it"
	"github.com/codahale/belvedere/pkg/belvedere/internal/waiter"
	"google.golang.org/api/cloudresourcemanager/v1"
	"gopkg.in/h2non/gock.v1"
)

func TestSetDMPerms(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://cloudresourcemanager.googleapis.com/v1/projects/my-project?alt=json&fields=projectNumber&prettyPrint=false").
		Reply(200).
		JSON(cloudresourcemanager.Project{
			ProjectNumber: 123456,
		})

	gock.New("https://cloudresourcemanager.googleapis.com/v1/projects/my-project:getIamPolicy?alt=json&prettyPrint=false").
		Reply(200).
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
		Reply(200).
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
		Reply(409)

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
		Reply(200).
		JSON(cloudresourcemanager.Policy{})

	ctx := waiter.WithInterval(context.TODO(), 10*time.Millisecond)
	if err := SetDMPerms(ctx, "my-project", false); err != nil {
		t.Fatal(err)
	}
}
