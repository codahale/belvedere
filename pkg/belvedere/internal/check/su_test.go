package check

import (
	"context"
	"testing"

	"github.com/codahale/belvedere/pkg/belvedere/internal/it"
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

	su, err := serviceusage.NewService(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	f := SU(context.TODO(), su, "op1")
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

	su, err := serviceusage.NewService(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	f := SU(context.TODO(), su, "op1")
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

	su, err := serviceusage.NewService(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	f := SU(context.TODO(), su, "op1")
	_, err = f()
	if err == nil {
		t.Fatal("should have returned an error")
	}

	expected := "{\"code\":500,\"message\":\"nope\"}"
	if actual := err.Error(); actual != expected {
		t.Errorf("Expected %q but got %q", expected, actual)
	}

}