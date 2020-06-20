package gcp

import (
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/codahale/belvedere/internal/assert"
	"google.golang.org/api/googleapi"
)

func TestModifyLoop_Success(t *testing.T) {
	n := 0
	err := ModifyLoop(10*time.Millisecond, 1*time.Second, func() error {
		n++
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "ModifyLoop count", 1, n)
}

func TestModifyLoop_PreconditionFailed(t *testing.T) {
	n := 0
	err := ModifyLoop(10*time.Millisecond, 1*time.Second, func() error {
		n++
		if n < 3 {
			return &googleapi.Error{Code: http.StatusPreconditionFailed}
		}
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "ModifyLoop count", 3, n)
}

func TestModifyLoop_InitialFailure(t *testing.T) {
	n := 0
	err := ModifyLoop(10*time.Millisecond, 1*time.Second, func() error {
		n++
		if n < 3 {
			return &googleapi.Error{Code: http.StatusConflict}
		}
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "ModifyLoop count", 3, n)
}

func TestModifyLoop_FinalFailure(t *testing.T) {
	n := 0
	err := ModifyLoop(10*time.Millisecond, 100*time.Millisecond, func() error {
		n++
		if n < 3 {
			return &googleapi.Error{Code: http.StatusConflict}
		}
		return nil
	})

	if err == nil || !strings.HasPrefix(err.Error(), "timeout after") {
		t.Fatal("bad error")
	}

	if n < 1 {
		t.Fatal("no loop runs")
	}
}

func TestModifyLoopFatalFailure(t *testing.T) {
	n := 0
	err := ModifyLoop(10*time.Millisecond, 100*time.Millisecond, func() error {
		n++
		return os.ErrClosed
	})

	if err != os.ErrClosed {
		t.Fatal(err)
	}

	assert.Equal(t, "ModifyLoop count", 1, n)
}
