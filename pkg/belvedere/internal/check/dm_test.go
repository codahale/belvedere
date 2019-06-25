package check

import (
	"context"
	"testing"

	"github.com/codahale/belvedere/pkg/belvedere/internal/it"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/deploymentmanager/v2"
	"gopkg.in/h2non/gock.v1"
)

func TestDMRunning(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://www.googleapis.com/deploymentmanager/v2/projects/example/global/operations/op1?alt=json&fields=status%2Cerror&prettyPrint=false").
		Reply(200).
		JSON(deploymentmanager.Operation{
			Status: "RUNNING",
		})

	f := DM(context.TODO(), "example", "op1")
	done, err := f()
	if err != nil {
		t.Fatal(err)
	}

	if done {
		t.Error("was done but shouldn't have been")
	}
}

func TestDMDone(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://www.googleapis.com/deploymentmanager/v2/projects/example/global/operations/op1?alt=json&fields=status%2Cerror&prettyPrint=false").
		Reply(200).
		JSON(deploymentmanager.Operation{
			Status: "DONE",
		})

	f := DM(context.TODO(), "example", "op1")
	done, err := f()
	if err != nil {
		t.Fatal(err)
	}

	if !done {
		t.Error("wasn't done but should have been")
	}
}

func TestDMError(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://www.googleapis.com/deploymentmanager/v2/projects/example/global/operations/op1?alt=json&fields=status%2Cerror&prettyPrint=false").
		Reply(200).
		JSON(deploymentmanager.Operation{
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
		})

	f := DM(context.TODO(), "example", "op1")
	_, err := f()
	if err == nil {
		t.Fatal("no error, but should have returned one")
	}

	expected := "{\"errors\":[{\"code\":\"ERR_BAD_NEWS\",\"location\":\"/downtown\",\"message\":\"here comes Mongo\"}]}"
	if actual := err.Error(); actual != expected {
		t.Error(cmp.Diff(actual, expected))
	}
}
