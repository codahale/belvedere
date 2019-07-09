package main

import (
	"encoding/csv"
	"os"

	"github.com/olekukonko/tablewriter"
	"golang.org/x/sys/unix"
)

func isTerminal() bool {
	_, err := unix.IoctlGetTermios(int(os.Stdout.Fd()), unix.TIOCGETA)
	return err == nil
}

func printTable(headers []string, rows [][]string) error {
	if isTerminal() {
		tw := tablewriter.NewWriter(os.Stdout)
		tw.SetAutoFormatHeaders(false)
		tw.SetAutoWrapText(false)
		tw.SetHeader(headers)
		tw.AppendBulk(rows)
		tw.Render()
	} else {
		cw := csv.NewWriter(os.Stdout)
		_ = cw.Write(headers)
		_ = cw.WriteAll(rows)
		cw.Flush()
	}
	return nil
}
