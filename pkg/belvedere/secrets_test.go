package belvedere

import (
	"context"
	"testing"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/internal/it"
	"github.com/codahale/belvedere/pkg/belvedere/internal/waiter"
	secretmanager "google.golang.org/api/secretmanager/v1beta1"
	"gopkg.in/h2non/gock.v1"
)

func TestGrantAppSecret(t *testing.T) {
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
	if err := GrantAppSecret(ctx, "my-project", "my-app", "my-secret", false); err != nil {
		t.Fatal(err)
	}
}

func TestRevokeAppSecret(t *testing.T) {
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
	if err := RevokeAppSecret(ctx, "my-project", "my-app", "my-secret", false); err != nil {
		t.Fatal(err)
	}
}
