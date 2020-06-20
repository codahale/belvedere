package check

import (
	"context"
	"net/http"
	"testing"

	"github.com/codahale/belvedere/internal/assert"
	"github.com/codahale/belvedere/pkg/belvedere/internal/it"
	"google.golang.org/api/option"
	"google.golang.org/api/serviceusage/v1"
	"gopkg.in/h2non/gock.v1"
)

func TestSURunning(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://serviceusage.googleapis.com/v1/op1?alt=json&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(serviceusage.Operation{
			Done: false,
		})

	su, err := serviceusage.NewService(context.Background(), option.WithHTTPClient(http.DefaultClient))
	if err != nil {
		t.Fatal(err)
	}

	f := SU(context.Background(), su, "op1")
	done, err := f()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "SU()", false, done)
}

func TestSUDone(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://serviceusage.googleapis.com/v1/op1?alt=json&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(serviceusage.Operation{
			Done: true,
		})

	su, err := serviceusage.NewService(context.Background(), option.WithHTTPClient(http.DefaultClient))
	if err != nil {
		t.Fatal(err)
	}

	f := SU(context.Background(), su, "op1")
	done, err := f()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "SU()", true, done)
}

func TestSUError(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://serviceusage.googleapis.com/v1/op1?alt=json&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(serviceusage.Operation{
			Done: true,
			Error: &serviceusage.Status{
				Code:    500,
				Message: "nope",
			},
		})

	su, err := serviceusage.NewService(context.Background(), option.WithHTTPClient(http.DefaultClient))
	if err != nil {
		t.Fatal(err)
	}

	f := SU(context.Background(), su, "op1")
	_, err = f()
	if err == nil {
		t.Fatal("should have returned an error")
	}

	want := `operation failed: {"code":500,"message":"nope"}`
	assert.Equal(t, "SU() error", want, err.Error())
}
