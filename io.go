package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"syscall"

	"github.com/alecthomas/kong"
	"golang.org/x/crypto/ssh/terminal"
)

type FileContentFlag []byte

func (f *FileContentFlag) Decode(ctx *kong.DecodeContext) error {
	var filename string
	err := ctx.Scan.PopValueInto("filename", &filename)
	if err != nil {
		return err
	}

	if filename == "-" {
		if terminal.IsTerminal(syscall.Stdin) {
			return fmt.Errorf("can't read from stdin")
		}

		defer func() { _ = os.Stdin.Close() }()

		data, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		*f = data
	} else {
		filename = kong.ExpandPath(filename)
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("failed to open %q: %w", filename, err)
		}
		*f = data
	}
	return nil
}
