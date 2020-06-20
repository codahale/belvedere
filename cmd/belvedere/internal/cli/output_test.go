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

	want := `
+-------+-----------+
| Name  | Regular   |
+-------+-----------+
| one   |       two |
| three | four five |
+-------+-----------+
`
	got := "\n" + buf.String()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Table output mismatch (-want +got):\n%s", diff)
	}
}

func TestCSVOutput_Print(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	output := &csvOutput{tableOutput: tableOutput{w: buf}}
	if err := output.Print([]example{
		{
			Name:  "one",
			Weird: "two",
		},
	}); err != nil {
		t.Fatal(err)
	}

	want := `
Name,Regular
one,two
`
	got := "\n" + buf.String()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("CSV output mismatch (-want +got):\n%s", diff)
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

	want := `
{"Name":"one","Weird":"two"}
`
	got := "\n" + buf.String()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("JSON output mismatch (-want +got):\n%s", diff)
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

	want := `
{
  "Name": "one",
  "Weird": "two"
}
`
	got := "\n" + buf.String()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Pretty JSON output mismatch (-want +got):\n%s", diff)
	}
}

func TestYamlOutput_Print(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	output := &yamlOutput{w: buf}
	if err := output.Print([]example{
		{
			Name:  "one",
			Weird: "two",
		},
	}); err != nil {
		t.Fatal(err)
	}

	want := `
Name: one
Weird: two

`
	got := "\n" + buf.String()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("YAML output mismatch (-want +got):\n%s", diff)
	}
}
