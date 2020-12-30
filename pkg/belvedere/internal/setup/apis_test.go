package setup

import (
	"context"
	"testing"
	"time"

	"github.com/codahale/belvedere/internal/httpmock"
	"google.golang.org/api/option"
	"google.golang.org/api/serviceusage/v1"
)

func TestManager_EnableAPIs(t *testing.T) {
	t.Parallel()

	srv := httpmock.NewServer(t)
	defer srv.Finish()

	srv.Expect(`/v1/projects/my-project/services:batchEnable?alt=json&prettyPrint=false`,
		httpmock.ReqJSON(serviceusage.BatchEnableServicesRequest{
			ServiceIds: requiredServices[:20],
		}),
		httpmock.RespJSON(serviceusage.Operation{
			Name: "op1",
		}))

	srv.Expect(`/v1/op1?alt=json&prettyPrint=false`,
		httpmock.RespJSON(serviceusage.Operation{
			Done: true,
		}))

	srv.Expect(`/v1/projects/my-project/services:batchEnable?alt=json&prettyPrint=false`,
		httpmock.ReqJSON(serviceusage.BatchEnableServicesRequest{
			ServiceIds: requiredServices[20:],
		}),
		httpmock.RespJSON(serviceusage.Operation{
			Name: "op2",
		}))

	srv.Expect(`/v1/op2?alt=json&prettyPrint=false`,
		httpmock.RespJSON(serviceusage.Operation{
			Done: true,
		}))

	su, err := serviceusage.NewService(
		context.Background(),
		option.WithEndpoint(srv.URL()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	m := &service{su: su}

	if err := m.EnableAPIs(context.Background(), "my-project", false, 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}
