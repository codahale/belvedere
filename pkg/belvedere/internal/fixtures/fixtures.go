package fixtures

import (
	"bytes"
	"io/ioutil"
	"os"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Compare validates that the contents of the given file are the same as the given bytes. If the
// OVERWRITE environment variable is set to TRUE, the given bytes are written to the file first.
func Compare(t *testing.T, filename string, actual []byte) {
	if ok, _ := strconv.ParseBool(os.Getenv("OVERWRITE")); ok {
		t.Logf("overwriting %s", filename)
		err := ioutil.WriteFile(filename, actual, 0644) //nolint:gosec
		if err != nil {
			t.Fatal(err)
		}
	}

	expected, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(expected, actual) {
		t.Error(cmp.Diff(string(expected), string(actual)))
	}
}
