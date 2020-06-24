package check

import (
	"context"
	"net/http"
	"testing"

	"github.com/codahale/belvedere/internal/assert"
	"google.golang.org/api/option"
	"google.golang.org/api/serviceusage/v1"
	"gopkg.in/h2non/gock.v1"
)

func TestSU(t *testing.T) {
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
			defer gock.Off()

			gock.New("https://serviceusage.googleapis.com/v1/op1?alt=json&prettyPrint=false").
				Reply(http.StatusOK).
				JSON(testCase.op)

			su, err := serviceusage.NewService(
				context.Background(),
				option.WithHTTPClient(http.DefaultClient),
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
