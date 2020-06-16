package cli

import (
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/spf13/afero"
)

func NewFs() afero.Fs {
	return afero.Fs(stdInFs{Fs: afero.NewOsFs()})
}

const StdIn = "-"

type stdInFs struct {
	afero.Fs
}

func (fs stdInFs) Open(name string) (afero.File, error) {
	if name == StdIn {
		if isTerminal(os.Stdin.Fd()) {
			return nil, fmt.Errorf("can't read from stdin")
		}
		return os.Stdin, nil
	}
	return fs.Fs.Open(name)
}

func isTerminal(fd uintptr) bool {
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}
