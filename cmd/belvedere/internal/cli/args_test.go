package cli

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestArgs_String(t *testing.T) {
	args := &args{args: []string{"one", "two"}}

	want, got := "two", args.String(1)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("String() mismatch (-want +got):\n%s", diff)
	}
}

func TestArgs_String_Default(t *testing.T) {
	args := &args{args: []string{"one", "two"}}

	want, got := "", args.String(20)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("String() mismatch (-want +got):\n%s", diff)
	}
}

func TestArgs_File(t *testing.T) {
	args := &args{
		args: []string{"example.txt"},
	}

	want := []byte("one\n")
	got, err := args.File(0)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("File() mismatch (-want +got):\n%s", diff)
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

	want := []byte("one\n")
	got, err := args.File(0)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("File() mismatch (-want +got):\n%s", diff)
	}
}
