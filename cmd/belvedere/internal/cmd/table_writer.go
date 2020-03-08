package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"reflect"
	"syscall"
	"time"

	"github.com/olekukonko/tablewriter"
	"golang.org/x/crypto/ssh/terminal"
)

type TableWriter interface {
	Print(v interface{}) error
}

func NewTableWriter(csv bool) TableWriter {
	return &tableWriter{csv: csv}
}

type tableWriter struct {
	csv bool
}

func (w *tableWriter) Print(i interface{}) error {
	t := reflect.TypeOf(i)
	if t.Kind() != reflect.Slice {
		return nil
	}

	t = t.Elem()
	if t.Kind() != reflect.Struct {
		return nil
	}

	var headers []string
	for i := 0; i < t.NumField(); i++ {
		s := t.Field(i).Tag.Get("table")
		if s == "" {
			s = t.Field(i).Name
		}
		headers = append(headers, s)
	}

	var rows [][]string
	iv := reflect.ValueOf(i)
	for i := 0; i < iv.Len(); i++ {
		var row []string
		ev := iv.Index(i)
		for j := range headers {
			f := ev.Field(j)

			if t, ok := f.Interface().(time.Time); ok {
				row = append(row, t.Format(time.Stamp))
			} else {
				row = append(row, fmt.Sprint(f.Interface()))
			}
		}
		rows = append(rows, row)
	}

	if terminal.IsTerminal(syscall.Stdout) && !w.csv {
		tw := tablewriter.NewWriter(os.Stdout)
		tw.SetAutoFormatHeaders(false)
		tw.SetAutoWrapText(false)
		tw.SetHeader(headers)
		tw.AppendBulk(rows)
		tw.Render()
	} else {
		cw := csv.NewWriter(os.Stdout)
		if err := cw.Write(headers); err != nil {
			return err
		}
		if err := cw.WriteAll(rows); err != nil {
			return err
		}
		cw.Flush()
	}
	return nil
}