package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"

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

	path = expandPath(path)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %q: %w", path, err)
	}
	return data, nil
}

func expandPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	if strings.HasPrefix(path, "~/") {
		u, err := user.Current()
		if err != nil {
			return path
		}
		return filepath.Join(u.HomeDir, path[2:])
	}
	abspath, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return abspath
}
