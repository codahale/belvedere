package cli

import (
	"bytes"
	"testing"

	"github.com/codahale/belvedere/internal/assert"
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
	assert.Equal(t, "Table", want, got)
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

	assert.Equal(t, "CSV", want, got)
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

	assert.Equal(t, "JSON", want, got)
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

	assert.Equal(t, "Pretty JSON", want, got)
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

	assert.Equal(t, "YAML", want, got)
}
