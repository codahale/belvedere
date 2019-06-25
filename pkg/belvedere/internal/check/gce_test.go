package check

import (
	"context"
	"testing"

	"github.com/codahale/belvedere/pkg/belvedere/internal/it"
	compute "google.golang.org/api/compute/v0.beta"
	"gopkg.in/h2non/gock.v1"
)

func TestGCERunning(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gce, err := compute.NewService(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	gock.New("https://www.googleapis.com/compute/beta/projects/example/global/operations/op1?alt=json&fields=status%2Cerror&prettyPrint=false").
		Reply(200).
		JSON(compute.Operation{
			Status: "RUNNING",
		})

	f := GCE(context.TODO(), gce, "example", "op1")
	done, err := f()
	if err != nil {
		t.Fatal(err)
	}

	if done {
		t.Error("shouldn't have been done, but was")
	}
}

func TestGCEDone(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gce, err := compute.NewService(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	gock.New("https://www.googleapis.com/compute/beta/projects/example/global/operations/op1?alt=json&fields=status%2Cerror&prettyPrint=false").
		Reply(200).
		JSON(compute.Operation{
			Status: "DONE",
		})

	f := GCE(context.TODO(), gce, "example", "op1")
	done, err := f()
	if err != nil {
		t.Fatal(err)
	}

	if !done {
		t.Error("should have been done, but wasn't")
	}
}

func TestGCEError(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gce, err := compute.NewService(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	gock.New("https://www.googleapis.com/compute/beta/projects/example/global/operations/op1?alt=json&fields=status%2Cerror&prettyPrint=false").
		Reply(200).
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

	f := GCE(context.TODO(), gce, "example", "op1")
	_, err = f()
	if err == nil {
		t.Fatal("should have returned an error")
	}

	expected := "{\"errors\":[{\"code\":\"ERR_MAGIC_HAT\",\"location\":\"/great-hall\",\"message\":\"Bad personality test\"}]}"
	if actual := err.Error(); actual != expected {
		t.Errorf("Expected %q but got %q", expected, actual)
	}
}
