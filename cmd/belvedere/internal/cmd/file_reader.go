package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"syscall"

	"github.com/alecthomas/kong"
	"golang.org/x/crypto/ssh/terminal"
)

type FileReader interface {
	Read(path string) ([]byte, error)
}

func NewFileReader() FileReader {
	return &fileReader{}
}

type fileReader struct {
}

func (f *fileReader) Read(path string) ([]byte, error) {
	if path == "-" {
		if terminal.IsTerminal(syscall.Stdin) {
			return nil, fmt.Errorf("can't read from stdin")
		}

		defer func() { _ = os.Stdin.Close() }()

		data, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("error reading from stdin: %w", err)
		}
		return data, nil
	}

	path = kong.ExpandPath(path)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %q: %w", path, err)
	}
	return data, nil
}
