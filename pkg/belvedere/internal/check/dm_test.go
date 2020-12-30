package check

import (
	"context"
	"testing"

	"github.com/codahale/belvedere/internal/assert"
	"github.com/codahale/belvedere/internal/httpmock"
	"google.golang.org/api/deploymentmanager/v2"
	"google.golang.org/api/option"
)

func TestDM(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		op     deploymentmanager.Operation
		done   bool
		errMsg string
	}{
		{
			name: "running",
			op: deploymentmanager.Operation{
				Status: "RUNNING",
			},
			done: false,
		},
		{
			name: "done",
			op: deploymentmanager.Operation{
				Status: "DONE",
			},
			done: true,
		},
		{
			name: "error",
			op: deploymentmanager.Operation{
				Status: "DONE",
				Error: &deploymentmanager.OperationError{
					Errors: []*deploymentmanager.OperationErrorErrors{
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

			srv.Expect(`/deploymentmanager/v2/projects/example/global/operations/op1?`+
				`alt=json&fields=status%2Cerror&prettyPrint=false`,
				httpmock.RespJSON(testCase.op))

			dm, err := deploymentmanager.NewService(
				context.Background(),
				option.WithEndpoint(srv.URL()),
				option.WithoutAuthentication(),
			)
			if err != nil {
				t.Fatal(err)
			}

			done, err := DM(context.Background(), dm, "example", "op1")()

			assert.Equal(t, "done", testCase.done, done)

			errMsg := ""
			if err != nil {
				errMsg = err.Error()
			}

			assert.Equal(t, "errMsg", testCase.errMsg, errMsg)
		})
	}
}
