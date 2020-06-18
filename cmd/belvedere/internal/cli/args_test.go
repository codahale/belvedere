package cli

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestArgs_String(t *testing.T) {
	args := &args{args: []string{"one", "two"}}

	expected, actual := "two", args.String(1)
	if !cmp.Equal(expected, actual) {
		t.Fatal(cmp.Diff(expected, actual))
	}
}

func TestArgs_String_Default(t *testing.T) {
	args := &args{args: []string{"one", "two"}}

	expected, actual := "", args.String(20)
	if !cmp.Equal(expected, actual) {
		t.Fatal(cmp.Diff(expected, actual))
	}
}

func TestArgs_File(t *testing.T) {
	args := &args{
		args: []string{"example.txt"},
	}

	expected := []byte("one\n")
	actual, err := args.File(0)
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(expected, actual) {
		t.Fatal(cmp.Diff(expected, actual))
	}
}

func TestArgs_File_Default(t *testing.T) {
	f, err := os.Open("example.txt")
	if err != nil {
		t.Fatal(err)
	}
	args := &args{
		stdin: f,
	}

	expected := []byte("one\n")
	actual, err := args.File(0)
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(expected, actual) {
		t.Fatal(cmp.Diff(expected, actual))
	}
}
