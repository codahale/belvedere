package gcp

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPaginate(t *testing.T) {
	var tokens []string
	if err := Paginate(func(s string) (string, error) {
		if len(tokens) == 3 {
			return "", nil
		}
		tokens = append(tokens, s)
		return "woo", nil
	}); err != nil {
		t.Fatal(err)
	}

	expected, actual := []string{"", "woo", "woo"}, tokens
	if !cmp.Equal(expected, actual) {
		t.Fatal(cmp.Diff(expected, actual))
	}
}

func TestPaginateErr(t *testing.T) {
	expected := os.ErrClosed
	actual := Paginate(func(s string) (string, error) {
		return "", expected
	})

	if expected != actual {
		t.Fatal(cmp.Diff(expected, actual))
	}
}
