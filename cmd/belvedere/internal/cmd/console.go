// +build darwin dragonfly freebsd linux netbsd openbsd

package cmd

import (
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

func isStdInReadable() bool {
	return !terminal.IsTerminal(syscall.Stdin)
}
