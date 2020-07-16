package assert

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"testing"
)

// EqualFixture validates that the contents of the given file are the same as the given bytes. If the
// OVERWRITE environment variable is set to TRUE, the given bytes are written to the file first.
func EqualFixture(t *testing.T, name, filename string, got []byte) {
	t.Helper()

	if ok, _ := strconv.ParseBool(os.Getenv("OVERWRITE")); ok {
		t.Logf("overwriting %s", filename)

		err := ioutil.WriteFile(filename, got, 0o644) //nolint:gosec // not used in main code
		if err != nil {
			t.Fatal(err)
		}
	}

	want, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	Equal(t, fmt.Sprintf("%s/%s", name, filename), want, got)
}
