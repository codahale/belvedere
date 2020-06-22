package setup

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/codahale/belvedere/internal/it"
	"google.golang.org/api/option"
	"google.golang.org/api/serviceusage/v1"
	"gopkg.in/h2non/gock.v1"
)

func TestManager_EnableAPIs(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://serviceusage.googleapis.com/v1/projects/my-project/services:batchEnable?alt=json&prettyPrint=false").
		JSON(serviceusage.BatchEnableServicesRequest{
			ServiceIds: requiredServices[:20],
		}).
		Reply(http.StatusOK).
		JSON(serviceusage.Operation{
			Name: "op1",
		})

	gock.New("https://serviceusage.googleapis.com/v1/op1?alt=json&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(serviceusage.Operation{
			Done: true,
		})

	gock.New("https://serviceusage.googleapis.com/v1/projects/my-project/services:batchEnable?alt=json&prettyPrint=false").
		JSON(serviceusage.BatchEnableServicesRequest{
			ServiceIds: requiredServices[20:],
		}).
		Reply(http.StatusOK).
		JSON(serviceusage.Operation{
			Name: "op2",
		})

	gock.New("https://serviceusage.googleapis.com/v1/op2?alt=json&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(serviceusage.Operation{
			Done: true,
		})

	su, err := serviceusage.NewService(context.Background(), option.WithHTTPClient(http.DefaultClient))
	if err != nil {
		t.Fatal(err)
	}

	m := &service{su: su}

	if err := m.EnableAPIs(context.Background(), "my-project", false, 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}
