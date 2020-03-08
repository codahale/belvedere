package cmd

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestWrap(t *testing.T) {
	expected := `This is a single paragraph.

Here's another paragraph which is formatted oddly.

Again, we conclude that there is no better way of solving this issue than
parsing the text into paragraphs and wrapping each.`
	actual := Wrap(`This is a single paragraph.

Here's another paragraph
which is formatted
oddly.

Again, we conclude that there is no better way of solving this issue than parsing the text into paragraphs and wrapping each.`)

	if !cmp.Equal(expected, actual) {
		t.Fatal(cmp.Diff(expected, actual))
	}
}
