package check

import (
	"context"
	"testing"

	"github.com/codahale/belvedere/internal/assert"
	"github.com/codahale/belvedere/internal/httpmock"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

func TestGCE(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		op     compute.Operation
		done   bool
		errMsg string
	}{
		{
			name: "running",
			op: compute.Operation{
				Status: "RUNNING",
			},
			done: false,
		},
		{
			name: "done",
			op: compute.Operation{
				Status: "DONE",
			},
			done: true,
		},
		{
			name: "error",
			op: compute.Operation{
				Status: "DONE",
				Error: &compute.OperationError{
					Errors: []*compute.OperationErrorErrors{
						{
							Code:     "ERR_BAD_NEWS",
							Location: "/downtown",
							Message:  "here comes Mongo",
						},
					},
				},
			},
			done:   false,
			errMsg: `operation failed: {"errors":[{"code":"ERR_BAD_NEWS","location":"/downtown","message":"here comes Mongo"}]}`,
		},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			srv := httpmock.NewServer(t)
			defer srv.Finish()

			srv.Expect(`/projects/example/global/operations/op1?alt=json&fields=status%2Cerror&prettyPrint=false`,
				httpmock.RespJSON(testCase.op))

			gce, err := compute.NewService(
				context.Background(),
				option.WithEndpoint(srv.URL()),
				option.WithoutAuthentication(),
			)
			if err != nil {
				t.Fatal(err)
			}

			done, err := GCE(context.Background(), gce, "example", "op1")()

			assert.Equal(t, "done", testCase.done, done)

			errMsg := ""
			if err != nil {
				errMsg = err.Error()
			}

			assert.Equal(t, "errMsg", testCase.errMsg, errMsg)
		})
	}
}
