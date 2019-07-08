package internal

import (
	"encoding/csv"
	"io"
	"os"

	"github.com/olekukonko/tablewriter"
	"golang.org/x/sys/unix"
)

func isTerminal() bool {
	_, err := unix.IoctlGetTermios(int(os.Stdout.Fd()), unix.TIOCGETA)
	return err == nil
}

func PrintTable(w io.Writer, rows [][]string, headers ...string) error {
	if isTerminal() {
		tw := tablewriter.NewWriter(w)
		tw.SetAutoFormatHeaders(false)
		tw.SetHeader(headers)
		tw.AppendBulk(rows)
		tw.Render()
	} else {
		cw := csv.NewWriter(w)
		_ = cw.Write(headers)
		_ = cw.WriteAll(rows)
		cw.Flush()
	}
	return nil
}
