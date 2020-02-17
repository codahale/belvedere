package gcp

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/googleapi"
)

func TestModifyLoopSuccess(t *testing.T) {
	n := 0
	err := ModifyLoop(10*time.Millisecond, 1*time.Second, func() error {
		n += 1
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	if expected, actual := 1, n; expected != actual {
		t.Fatal(cmp.Diff(expected, actual))
	}
}

func TestModifyLoopInitialFailure(t *testing.T) {
	n := 0
	err := ModifyLoop(10*time.Millisecond, 1*time.Second, func() error {
		n += 1
		if n < 3 {
			return &googleapi.Error{Code: 409}
		}
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	if expected, actual := 3, n; expected != actual {
		t.Fatal(cmp.Diff(expected, actual))
	}
}

func TestModifyLoopFinalFailure(t *testing.T) {
	n := 0
	err := ModifyLoop(10*time.Millisecond, 100*time.Millisecond, func() error {
		n += 1
		if n < 3 {
			return &googleapi.Error{Code: 409}
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
		n += 1
		return os.ErrClosed
	})

	if err != os.ErrClosed {
		t.Fatal(err)
	}

	if expected, actual := 1, n; expected != actual {
		t.Fatal(cmp.Diff(expected, actual))
	}
}
