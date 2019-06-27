package check

import (
	"context"
	"testing"

	"github.com/codahale/belvedere/pkg/belvedere/internal/it"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/serviceusage/v1"
	"gopkg.in/h2non/gock.v1"
)

func TestSURunning(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://serviceusage.googleapis.com/v1/op1?alt=json&fields=status%2Cerror&prettyPrint=false").
		Reply(200).
		JSON(serviceusage.Operation{
			Done: false,
		})

	f := SU(context.TODO(), "op1")
	done, err := f()
	if err != nil {
		t.Fatal(err)
	}

	if done {
		t.Error("shouldn't have been done but was")
	}
}

func TestSUDone(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://serviceusage.googleapis.com/v1/op1?alt=json&fields=status%2Cerror&prettyPrint=false").
		Reply(200).
		JSON(serviceusage.Operation{
			Done: true,
		})

	f := SU(context.TODO(), "op1")
	done, err := f()
	if err != nil {
		t.Fatal(err)
	}

	if !done {
		t.Error("should have been done but wasn't")
	}
}

func TestSUError(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://serviceusage.googleapis.com/v1/op1?alt=json&fields=status%2Cerror&prettyPrint=false").
		Reply(200).
		JSON(serviceusage.Operation{
			Done: true,
			Error: &serviceusage.Status{
				Code:    500,
				Message: "nope",
			},
		})

	f := SU(context.TODO(), "op1")
	_, err := f()
	if err == nil {
		t.Fatal("should have returned an error")
	}

	expected := "operation failed: {\"code\":500,\"message\":\"nope\"}"
	if actual := err.Error(); expected != actual {
		t.Error(cmp.Diff(expected, actual))
	}
}
