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

func (i *args) String(idx int) string {
	if len(i.args) > idx {
		return i.args[idx]
	}
	return ""
}

var errStdinUnreadable = fmt.Errorf("can't read from stdin")

func (i *args) File(idx int) ([]byte, error) {
	if len(i.args) > idx {
		return ioutil.ReadFile(i.args[idx])
	}

	if isTerminal(i.stdin) {
		return nil, errStdinUnreadable
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
