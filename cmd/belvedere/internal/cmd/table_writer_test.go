package cmd

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type example struct {
	Name  string
	Weird string `table:"Regular"`
}

func TestTableWriter_ASCII(t *testing.T) {
	out := bytes.NewBuffer(nil)
	tw := NewTableWriter(out, false)
	if err := tw.Print([]example{
		{
			Name:  "one",
			Weird: "two",
		},
	}); err != nil {
		t.Fatal(err)
	}

	expected := `
+------+---------+
| Name | Regular |
+------+---------+
| one  | two     |
+------+---------+
`
	actual := "\n" + out.String()

	if !cmp.Equal(expected, actual) {
		t.Fatal(cmp.Diff(expected, actual))
	}
}

func TestTableWriter_CSV(t *testing.T) {
	out := bytes.NewBuffer(nil)
	tw := NewTableWriter(out, true)
	if err := tw.Print([]example{
		{
			Name:  "one",
			Weird: "two",
		},
	}); err != nil {
		t.Fatal(err)
	}

	expected := `
Name,Regular
one,two
`
	actual := "\n" + out.String()

	if !cmp.Equal(expected, actual) {
		t.Fatal(cmp.Diff(expected, actual))
	}
}
