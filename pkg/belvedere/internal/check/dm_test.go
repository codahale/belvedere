package check

import (
	"context"
	"net/http"
	"testing"

	"github.com/codahale/belvedere/pkg/belvedere/internal/it"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/deploymentmanager/v2"
	"google.golang.org/api/option"
	"gopkg.in/h2non/gock.v1"
)

func TestDMRunning(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://www.googleapis.com/deploymentmanager/v2/projects/example/global/operations/op1?alt=json&fields=status%2Cerror&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(deploymentmanager.Operation{
			Status: "RUNNING",
		})

	dm, err := deploymentmanager.NewService(context.Background(), option.WithHTTPClient(http.DefaultClient))
	if err != nil {
		t.Fatal(err)
	}

	f := DM(context.Background(), dm, "example", "op1")
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
		Reply(http.StatusOK).
		JSON(deploymentmanager.Operation{
			Status: "DONE",
		})

	dm, err := deploymentmanager.NewService(context.Background(), option.WithHTTPClient(http.DefaultClient))
	if err != nil {
		t.Fatal(err)
	}

	f := DM(context.Background(), dm, "example", "op1")
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
		Reply(http.StatusOK).
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

	dm, err := deploymentmanager.NewService(context.Background(), option.WithHTTPClient(http.DefaultClient))
	if err != nil {
		t.Fatal(err)
	}

	f := DM(context.Background(), dm, "example", "op1")
	_, err = f()
	if err == nil {
		t.Fatal("no error, but should have returned one")
	}

	expected := "operation failed: {\"errors\":[{\"code\":\"ERR_BAD_NEWS\",\"location\":\"/downtown\",\"message\":\"here comes Mongo\"}]}"
	if actual := err.Error(); expected != actual {
		t.Error(cmp.Diff(expected, actual))
	}
}
