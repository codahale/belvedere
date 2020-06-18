package cli

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/mattn/go-isatty"
)

type Input interface {
	ReadFile(args []string, argIdx int) ([]byte, error)
}

type input struct {
	stdin io.Reader
}

func (i *input) ReadFile(args []string, argIdx int) ([]byte, error) {
	if len(args) > argIdx {
		return ioutil.ReadFile(args[argIdx])
	}

	if isTerminal(i.stdin) {
		return nil, fmt.Errorf("can't read from stdin")
	}

	if rc, ok := i.stdin.(io.ReadCloser); ok {
		defer func() { _ = rc.Close() }()
	}
	return ioutil.ReadAll(i.stdin)
}

func isTerminal(r io.Reader) bool {
	f, ok := r.(*os.File)
	if !ok {
		return false
	}
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}
