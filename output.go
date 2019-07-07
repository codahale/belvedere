package main

import (
	"bytes"

	"github.com/olekukonko/tablewriter"
)

func formatTable(headers []string, rows [][]string) string {
	b := bytes.NewBuffer(nil)
	tw := tablewriter.NewWriter(b)
	tw.SetAutoFormatHeaders(false)
	tw.SetHeader(headers)
	tw.AppendBulk(rows)
	tw.Render()
	return b.String()
}
