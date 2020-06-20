package fixtures

import (
	"io/ioutil"
	"os"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Compare validates that the contents of the given file are the same as the given bytes. If the
// OVERWRITE environment variable is set to TRUE, the given bytes are written to the file first.
func Compare(t *testing.T, filename string, got []byte) {
	if ok, _ := strconv.ParseBool(os.Getenv("OVERWRITE")); ok {
		t.Logf("overwriting %s", filename)
		err := ioutil.WriteFile(filename, got, 0644) //nolint:gosec
		if err != nil {
			t.Fatal(err)
		}
	}

	want, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("%s mismatch (-want +got):\n%s", filename, diff)
	}
}
