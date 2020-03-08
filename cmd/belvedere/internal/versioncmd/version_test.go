package versioncmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	"github.com/google/go-cmp/cmp"
)

func TestNew(t *testing.T) {
	out := bytes.NewBuffer(nil)
	command := New(&rootcmd.Config{}, out, "1.2.3", "abcdef", "2020-03-08", "coda")
	if err := command.ParseAndRun(context.Background(), nil); err != nil {
		t.Fatal(err)
	}

	actual := out.String()
	expected := `version: 1.2.3
commit: abcdef
built at: 2020-03-08
built by: coda
`

	if !cmp.Equal(expected, actual) {
		t.Fatal(cmp.Diff(expected, actual))
	}
}
