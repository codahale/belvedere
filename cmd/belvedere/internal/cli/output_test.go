package cli

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type example struct {
	Name  string
	Weird string `table:"Regular,ralign"`
}

func TestTableOutput_Print(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	output := &tableOutput{w: buf}
	if err := output.Print([]example{
		{
			Name:  "one",
			Weird: "two",
		},
		{
			Name:  "three",
			Weird: "four five",
		},
	}); err != nil {
		t.Fatal(err)
	}

	expected := `
+-------+-----------+
| Name  | Regular   |
+-------+-----------+
| one   |       two |
| three | four five |
+-------+-----------+
`
	actual := "\n" + buf.String()

	if !cmp.Equal(expected, actual) {
		t.Fatal(cmp.Diff(expected, actual))
	}
}

func TestCSVOutput_Print(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	output := &csvOutput{w: buf}
	if err := output.Print([]example{
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
	actual := "\n" + buf.String()

	if !cmp.Equal(expected, actual) {
		t.Fatal(cmp.Diff(expected, actual))
	}
}

func TestJSONOutput_Print(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	output := &jsonOutput{w: buf}
	if err := output.Print([]example{
		{
			Name:  "one",
			Weird: "two",
		},
	}); err != nil {
		t.Fatal(err)
	}

	expected := `
{"Name":"one","Weird":"two"}
`
	actual := "\n" + buf.String()

	if !cmp.Equal(expected, actual) {
		t.Fatal(cmp.Diff(expected, actual))
	}
}

func TestPrettyJSONOutput_Print(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	output := &prettyJSONOutput{w: buf}
	if err := output.Print([]example{
		{
			Name:  "one",
			Weird: "two",
		},
	}); err != nil {
		t.Fatal(err)
	}

	expected := `
{
  "Name": "one",
  "Weird": "two"
}
`
	actual := "\n" + buf.String()

	if !cmp.Equal(expected, actual) {
		t.Fatal(cmp.Diff(expected, actual))
	}
}
