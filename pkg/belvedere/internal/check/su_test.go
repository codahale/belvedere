package check

import (
	"context"
	"testing"

	"github.com/codahale/belvedere/internal/assert"
	"github.com/codahale/belvedere/internal/httpmock"
	"google.golang.org/api/option"
	"google.golang.org/api/serviceusage/v1"
)

func TestSU(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		op     serviceusage.Operation
		done   bool
		errMsg string
	}{
		{
			name: "running",
			op: serviceusage.Operation{
				Done: false,
			},
			done: false,
		},
		{
			name: "done",
			op: serviceusage.Operation{
				Done: true,
			},
			done: true,
		},
		{
			name: "error",
			op: serviceusage.Operation{
				Done: true,
				Error: &serviceusage.Status{
					Code:    500,
					Message: "nope",
				},
			},
			errMsg: `operation failed: {"code":500,"message":"nope"}`,
		},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			srv := httpmock.NewServer(t)
			defer srv.Finish()

			srv.Expect(`/v1/op1?alt=json&prettyPrint=false`,
				httpmock.RespJSON(testCase.op))

			su, err := serviceusage.NewService(
				context.Background(),
				option.WithEndpoint(srv.URL()),
				option.WithoutAuthentication(),
			)
			if err != nil {
				t.Fatal(err)
			}

			done, err := SU(context.Background(), su, "op1")()

			assert.Equal(t, "done", testCase.done, done)

			errMsg := ""
			if err != nil {
				errMsg = err.Error()
			}

			assert.Equal(t, "errMsg", testCase.errMsg, errMsg)
		})
	}
}
