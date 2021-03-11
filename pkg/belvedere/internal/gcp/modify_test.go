package gcp

import (
	"context"
	"errors"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/codahale/gubbins/assert"
	"google.golang.org/api/googleapi"
)

func TestModifyLoop_Success(t *testing.T) {
	t.Parallel()

	n := 0

	if err := ModifyLoop(10*time.Millisecond, 1*time.Second, func() error {
		n++
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "ModifyLoop count", 1, n)
}

func TestModifyLoop_PreconditionFailed(t *testing.T) {
	t.Parallel()

	n := 0

	if err := ModifyLoop(10*time.Millisecond, 1*time.Second, func() error {
		n++
		if n < 3 {
			return &googleapi.Error{Code: http.StatusPreconditionFailed}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "ModifyLoop count", 3, n)
}

func TestModifyLoop_InitialFailure(t *testing.T) {
	t.Parallel()

	n := 0

	if err := ModifyLoop(10*time.Millisecond, 1*time.Second, func() error {
		n++
		if n < 3 {
			return &googleapi.Error{Code: http.StatusConflict}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "ModifyLoop count", 3, n)
}

func TestModifyLoop_FinalFailure(t *testing.T) {
	t.Parallel()

	n := 0

	err := ModifyLoop(10*time.Millisecond, 100*time.Millisecond, func() error {
		n++
		if n < 3 {
			return &googleapi.Error{Code: http.StatusConflict}
		}
		return nil
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("bad error: %v", err)
	}

	if n < 1 {
		t.Fatal("no loop runs")
	}
}

func TestModifyLoopFatalFailure(t *testing.T) {
	t.Parallel()

	n := 0

	err := ModifyLoop(10*time.Millisecond, 100*time.Millisecond, func() error {
		n++
		return os.ErrClosed
	})
	if !errors.Is(err, os.ErrClosed) {
		t.Fatalf("bad error: %v", err)
	}

	assert.Equal(t, "ModifyLoop count", 1, n)
}
