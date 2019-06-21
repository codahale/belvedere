package belvedere

import (
	"errors"
	"io"
	"os"
)

func openPath(path string) (io.ReadCloser, error) {
	if path == "-" {
		return os.Stdin, nil
	}
	return os.Open(path)
}

var errUnimplemented = errors.New("unimplemented")
