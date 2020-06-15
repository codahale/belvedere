package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

type Output interface {
	Print(v interface{}) error
}

func NewOutput(w io.Writer, format string) (Output, error) {
	switch strings.ToLower(format) {
	case "table":
		return &tableOutput{w: w}, nil
	case "csv":
		return &csvOutput{w: w}, nil
	case "json":
		return &jsonOutput{w: w}, nil
	case "prettyjson":
		return &prettyJSONOutput{w: w}, nil
	default:
		return nil, fmt.Errorf("%q is not a valid format (must be one of: table, csv, json, prettyjson", format)
	}
}

type tableOutput struct {
	w io.Writer
}

func (o *tableOutput) Print(v interface{}) error {
	headers, cols, rows, err := toRows(v)
	if err != nil {
		return err
	}

	tw := table.NewWriter()
	tw.Style().Format.Header = text.FormatDefault
	tw.AppendHeader(headers)
	tw.SetColumnConfigs(cols)
	tw.AppendSeparator()
	tw.AppendRows(rows)
	_, err = fmt.Fprintln(o.w, tw.Render())
	return err
}

type csvOutput struct {
	w io.Writer
}

func (o *csvOutput) Print(v interface{}) error {
	headers, cols, rows, err := toRows(v)
	if err != nil {
		return err
	}

	tw := table.NewWriter()
	tw.AppendHeader(headers)
	tw.SetColumnConfigs(cols)
	tw.AppendSeparator()
	tw.AppendRows(rows)
	_, err = fmt.Fprintln(o.w, tw.RenderCSV())
	return err
}

type jsonOutput struct {
	w io.Writer
}

func (o *jsonOutput) Print(v interface{}) error {
	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Slice {
		return fmt.Errorf("not a slice of structs")
	}

	iv := reflect.ValueOf(v)
	for i := 0; i < iv.Len(); i++ {
		b, err := json.Marshal(iv.Index(i).Interface())
		if err != nil {
			return err
		}

		if _, err := fmt.Fprintf(o.w, "%s\n", string(b)); err != nil {
			return err
		}
	}
	return nil
}

type prettyJSONOutput struct {
	w io.Writer
}

func (o *prettyJSONOutput) Print(v interface{}) error {
	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Slice {
		return fmt.Errorf("not a slice of structs")
	}

	iv := reflect.ValueOf(v)
	for i := 0; i < iv.Len(); i++ {
		b, err := json.MarshalIndent(iv.Index(i).Interface(), "", "  ")
		if err != nil {
			return err
		}

		if _, err := fmt.Fprintf(o.w, "%s\n", string(b)); err != nil {
			return err
		}
	}
	return nil
}

func toRows(v interface{}) (table.Row, []table.ColumnConfig, []table.Row, error) {
	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Slice {
		return nil, nil, nil, fmt.Errorf("not a slice of structs")
	}

	t = t.Elem()
	if t.Kind() != reflect.Struct {
		return nil, nil, nil, fmt.Errorf("not a slice of structs")
	}

	var headers table.Row
	var cols []table.ColumnConfig
	for i := 0; i < t.NumField(); i++ {
		s := t.Field(i).Tag.Get("table")
		if s == "" {
			s = t.Field(i).Name
		}

		parts := strings.Split(s, ",")

		headers = append(headers, parts[0])
		if strings.Contains(s, ",ralign") {
			cols = append(cols, table.ColumnConfig{
				Name:  parts[0],
				Align: text.AlignRight,
			})
		}
	}

	var rows []table.Row
	iv := reflect.ValueOf(v)
	for i := 0; i < iv.Len(); i++ {
		var row table.Row
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

	return headers, cols, rows, nil
}
