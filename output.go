package main

import (
	"bytes"
	"encoding/csv"
	"os"

	"github.com/olekukonko/tablewriter"
	"golang.org/x/sys/unix"
)

func isTerminal() bool {
	_, err := unix.IoctlGetTermios(int(os.Stdout.Fd()), unix.TIOCGETA)
	return err == nil
}

func formatTable(headers []string, rows [][]string) string {
	b := bytes.NewBuffer(nil)
	if isTerminal() {
		tw := tablewriter.NewWriter(b)
		tw.SetAutoFormatHeaders(false)
		tw.SetHeader(headers)
		tw.AppendBulk(rows)
		tw.Render()
	} else {
		cw := csv.NewWriter(b)
		_ = cw.Write(headers)
		_ = cw.WriteAll(rows)
		cw.Flush()
	}
	return b.String()
}
