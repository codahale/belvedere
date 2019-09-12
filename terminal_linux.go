package main

import (
	"os"

	"golang.org/x/sys/unix"
)

func isTerminal() bool {
	_, err := unix.IoctlGetTermios(int(os.Stdout.Fd()), unix.TCGETA)
	return err == nil
}
