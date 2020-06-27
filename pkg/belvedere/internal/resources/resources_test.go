package resources

import (
	"testing"

	"github.com/codahale/belvedere/internal/assert"
)

func TestName(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		out  string
	}{
		{
			name: "base",
			in:   nil,
			out:  "belvedere",
		},
		{
			name: "app",
			in:   []string{"one"},
			out:  "belvedere-one",
		},
		{
			name: "release",
			in:   []string{"one", "two"},
			out:  "belvedere-one-two",
		},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, "Name()", testCase.out, Name(testCase.in...))
		})
	}
}