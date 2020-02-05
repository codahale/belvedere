package setup

import (
	"context"
	"testing"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/internal/it"
	"github.com/codahale/belvedere/pkg/belvedere/internal/waiter"
	"google.golang.org/api/serviceusage/v1"
	"gopkg.in/h2non/gock.v1"
)

func TestEnableAPIs(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://serviceusage.googleapis.com/v1/projects/my-project/services:batchEnable?alt=json&prettyPrint=false").
		JSON(serviceusage.BatchEnableServicesRequest{
			ServiceIds: requiredServices[:20],
		}).
		Reply(200).
		JSON(serviceusage.Operation{
			Name: "op1",
		})

	gock.New("https://serviceusage.googleapis.com/v1/op1?alt=json&prettyPrint=false").
		Reply(200).
		JSON(serviceusage.Operation{
			Done: true,
		})

	gock.New("https://serviceusage.googleapis.com/v1/projects/my-project/services:batchEnable?alt=json&prettyPrint=false").
		JSON(serviceusage.BatchEnableServicesRequest{
			ServiceIds: requiredServices[20:],
		}).
		Reply(200).
		JSON(serviceusage.Operation{
			Name: "op2",
		})

	gock.New("https://serviceusage.googleapis.com/v1/op2?alt=json&prettyPrint=false").
		Reply(200).
		JSON(serviceusage.Operation{
			Done: true,
		})

	ctx := waiter.WithInterval(context.TODO(), 10*time.Millisecond)
	if err := EnableAPIs(ctx, "my-project", false); err != nil {
		t.Fatal(err)
	}
}
