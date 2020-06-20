package check

import (
	"context"
	"net/http"
	"testing"

	"github.com/codahale/belvedere/internal/assert"
	"github.com/codahale/belvedere/pkg/belvedere/internal/it"
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	"gopkg.in/h2non/gock.v1"
)

func TestGCERunning(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://compute.googleapis.com/compute/v1/projects/example/global/operations/op1?alt=json&fields=status%2Cerror&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.Operation{
			Status: "RUNNING",
		})

	gce, err := compute.NewService(context.Background(), option.WithHTTPClient(http.DefaultClient))
	if err != nil {
		t.Fatal(err)
	}

	f := GCE(context.Background(), gce, "example", "op1")
	done, err := f()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "GCE()", false, done)
}

func TestGCEDone(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://compute.googleapis.com/compute/v1/projects/example/global/operations/op1?alt=json&fields=status%2Cerror&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.Operation{
			Status: "DONE",
		})

	gce, err := compute.NewService(context.Background(), option.WithHTTPClient(http.DefaultClient))
	if err != nil {
		t.Fatal(err)
	}

	f := GCE(context.Background(), gce, "example", "op1")
	done, err := f()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "GCE()", true, done)
}

func TestGCEError(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://compute.googleapis.com/compute/v1/projects/example/global/operations/op1?alt=json&fields=status%2Cerror&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.Operation{
			Status: "DONE",
			Error: &compute.OperationError{
				Errors: []*compute.OperationErrorErrors{
					{
						Code:     "ERR_MAGIC_HAT",
						Location: "/great-hall",
						Message:  "Bad personality test",
					},
				},
			},
		})

	gce, err := compute.NewService(context.Background(), option.WithHTTPClient(http.DefaultClient))
	if err != nil {
		t.Fatal(err)
	}

	f := GCE(context.Background(), gce, "example", "op1")
	_, err = f()
	if err == nil {
		t.Fatal("should have returned an error")
	}

	want := `operation failed: {"errors":[{"code":"ERR_MAGIC_HAT","location":"/great-hall","message":"Bad personality test"}]}`
	assert.Equal(t, "GCE() error", want, err.Error())
}
