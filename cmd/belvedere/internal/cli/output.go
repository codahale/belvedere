package cli

import (
	"encoding/csv"
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/olekukonko/tablewriter"
)

type Output interface {
	Print(v interface{}) error
}

func NewOutput(w io.Writer, csv bool) Output {
	return &output{csv: csv, w: w}
}

type output struct {
	csv bool
	w   io.Writer
}

func (w *output) Print(i interface{}) error {
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

	if w.csv {
		cw := csv.NewWriter(w.w)
		if err := cw.Write(headers); err != nil {
			return err
		}
		if err := cw.WriteAll(rows); err != nil {
			return err
		}
		cw.Flush()
	} else {
		tw := tablewriter.NewWriter(w.w)
		tw.SetAutoFormatHeaders(false)
		tw.SetAutoWrapText(false)
		tw.SetHeader(headers)
		tw.AppendBulk(rows)
		tw.Render()
	}
	return nil
}
