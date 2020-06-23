package cli

import (
	"os"
	"testing"

	"github.com/codahale/belvedere/internal/assert"
)

func TestArgs_String(t *testing.T) {
	args := &args{args: []string{"one", "two"}}

	assert.Equal(t, "String()", "two", args.String(1))
}

func TestArgs_String_Default(t *testing.T) {
	args := &args{args: []string{"one", "two"}}

	assert.Equal(t, "String()", "", args.String(20))
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

	assert.Equal(t, "File()", want, got)
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

	assert.Equal(t, "File()", want, got)
}
