package cli

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/mattn/go-isatty"
)

type Args interface {
	String(idx int) string
	File(idx int) ([]byte, error)
}

type args struct {
	args  []string
	stdin io.Reader
}

func (a *args) String(idx int) string {
	if len(a.args) > idx {
		return a.args[idx]
	}

	return ""
}

var errStdinUnreadable = fmt.Errorf("can't read from stdin")

func (a *args) File(idx int) ([]byte, error) {
	if len(a.args) > idx {
		return ioutil.ReadFile(a.args[idx])
	}

	if isTerminal(a.stdin) {
		return nil, errStdinUnreadable
	}

	if rc, ok := a.stdin.(io.ReadCloser); ok {
		defer func() { _ = rc.Close() }()
	}

	return ioutil.ReadAll(a.stdin)
}

func isTerminal(r io.Reader) bool {
	f, ok := r.(*os.File)
	if !ok {
		return false
	}

	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}
