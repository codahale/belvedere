package cli

import (
	"bytes"
	"testing"

	"github.com/codahale/gubbins/assert"
)

type example struct {
	Name    string
	Weird   string `table:"Regular,ralign"`
	Strings []string
}

//nolint:gochecknoglobals // tests
var testData = []example{
	{
		Name:  "one",
		Weird: "two",
	},
	{
		Name:    "three",
		Weird:   "four five",
		Strings: []string{"one", "two"},
	},
}

func TestTableOutput_Print(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBuffer(nil)
	output := &tableOutput{w: buf}

	if err := output.Print(testData); err != nil {
		t.Fatal(err)
	}

	want := `
+-------+-----------+---------+
| Name  | Regular   | Strings |
+-------+-----------+---------+
| one   |       two |         |
| three | four five | one,two |
+-------+-----------+---------+
`
	got := "\n" + buf.String()
	assert.Equal(t, "Table", want, got)
}

func TestCSVOutput_Print(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBuffer(nil)
	output := &csvOutput{tableOutput: tableOutput{w: buf}}

	if err := output.Print(testData); err != nil {
		t.Fatal(err)
	}

	want := `
Name,Regular,Strings
one,two,
three,four five,"one\,two"
`
	got := "\n" + buf.String()

	assert.Equal(t, "CSV", want, got)
}

func TestJSONOutput_Print(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBuffer(nil)
	output := &jsonOutput{w: buf}

	if err := output.Print(testData); err != nil {
		t.Fatal(err)
	}

	want := `
{"Name":"one","Weird":"two","Strings":null}
{"Name":"three","Weird":"four five","Strings":["one","two"]}
`
	got := "\n" + buf.String()

	assert.Equal(t, "JSON", want, got)
}

func TestPrettyJSONOutput_Print(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBuffer(nil)
	output := &prettyJSONOutput{w: buf}

	if err := output.Print(testData); err != nil {
		t.Fatal(err)
	}

	want := `
{
  "Name": "one",
  "Weird": "two",
  "Strings": null
}
{
  "Name": "three",
  "Weird": "four five",
  "Strings": [
    "one",
    "two"
  ]
}
`
	got := "\n" + buf.String()

	assert.Equal(t, "Pretty JSON", want, got)
}

func TestYamlOutput_Print(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBuffer(nil)
	output := &yamlOutput{w: buf}

	if err := output.Print(testData); err != nil {
		t.Fatal(err)
	}

	want := `
---
Name: one
Strings: null
Weird: two
---
Name: three
Strings:
- one
- two
Weird: four five
`
	got := "\n" + buf.String()

	assert.Equal(t, "YAML", want, got)
}
