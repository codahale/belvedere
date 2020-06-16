package cli

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPathFromArgs(t *testing.T) {
	expected, actual := "-", PathFromArgs([]string{}, 2)
	if !cmp.Equal(expected, actual) {
		t.Fatal(cmp.Diff(expected, actual))
	}

	expected, actual = "three", PathFromArgs([]string{"one", "two", "three"}, 2)
	if !cmp.Equal(expected, actual) {
		t.Fatal(cmp.Diff(expected, actual))
	}

	expected, actual = "three", PathFromArgs([]string{"one", "two", "three", "four"}, 2)
	if !cmp.Equal(expected, actual) {
		t.Fatal(cmp.Diff(expected, actual))
	}
}
