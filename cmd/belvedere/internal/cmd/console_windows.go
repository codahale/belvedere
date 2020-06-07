package cmd

import (
	"syscall"

	"github.com/Azure/go-ansiterm/winterm"
)

func isStdInReadable() bool {
	_, e := winterm.GetConsoleMode(uintptr(syscall.Stdin))
	return e == nil
}
