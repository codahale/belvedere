package versioncmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	"github.com/google/go-cmp/cmp"
)

func TestVersion(t *testing.T) {
	out := bytes.NewBuffer(nil)
	command := New(&rootcmd.Config{}, out, "1.2.3", "abcdef", "2020-03-08", "coda", "go1.14")
	if err := command.ParseAndRun(context.Background(), nil); err != nil {
		t.Fatal(err)
	}

	actual := out.String()
	expected := `version:    1.2.3
commit:     abcdef
built at:   2020-03-08
built by:   coda
built with: go1.14
`

	if !cmp.Equal(expected, actual) {
		t.Fatal(cmp.Diff(expected, actual))
	}
}
