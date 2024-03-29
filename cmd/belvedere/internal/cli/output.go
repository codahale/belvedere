package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

type Output interface {
	Print(v interface{}) error
}

var errBadFormat = fmt.Errorf("format must be one of: table, csv, json, prettyjson, yaml")

func NewOutput(w io.Writer, format string) (Output, error) {
	switch strings.ToLower(format) {
	case "table":
		return &tableOutput{w: w}, nil
	case "csv":
		return &csvOutput{tableOutput: tableOutput{w: w}}, nil
	case "json":
		return &jsonOutput{w: w}, nil
	case "prettyjson":
		return &prettyJSONOutput{w: w}, nil
	case "yaml":
		return &yamlOutput{w: w}, nil
	default:
		return nil, errBadFormat
	}
}

type tableOutput struct {
	w io.Writer
}

func (o *tableOutput) Print(v interface{}) error {
	tw, err := o.buildWriter(v)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(o.w, tw.Render())

	return err
}

func (o *tableOutput) buildWriter(v interface{}) (table.Writer, error) {
	headers, cols, rows, err := toRows(v)
	if err != nil {
		return nil, err
	}

	tw := table.NewWriter()
	tw.Style().Format.Header = text.FormatDefault
	tw.AppendHeader(headers)
	tw.SetColumnConfigs(cols)
	tw.AppendSeparator()
	tw.AppendRows(rows)

	return tw, err
}

type csvOutput struct {
	tableOutput
}

func (o *csvOutput) Print(v interface{}) error {
	tw, err := o.buildWriter(v)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(o.w, tw.RenderCSV())

	return err
}

type jsonOutput struct {
	w io.Writer
}

func (o *jsonOutput) Print(v interface{}) error {
	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Slice {
		return errBadInput
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
		return errBadInput
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

type yamlOutput struct {
	w io.Writer
}

func (o *yamlOutput) Print(v interface{}) error {
	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Slice {
		return errBadInput
	}

	iv := reflect.ValueOf(v)
	for i := 0; i < iv.Len(); i++ {
		b, err := yaml.Marshal(iv.Index(i).Interface())
		if err != nil {
			return err
		}

		if _, err := fmt.Fprintf(o.w, "---\n%s", string(b)); err != nil {
			return err
		}
	}

	return nil
}

var errBadInput = fmt.Errorf("not a slice of structs")

func toRows(v interface{}) (table.Row, []table.ColumnConfig, []table.Row, error) {
	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Slice {
		return nil, nil, nil, errBadInput
	}

	t = t.Elem()
	if t.Kind() != reflect.Struct {
		return nil, nil, nil, errBadInput
	}

	headers, cols := collectHeaders(t)
	rows := collectRows(v, headers)

	return headers, cols, rows, nil
}

func collectRows(v interface{}, headers table.Row) []table.Row {
	iv := reflect.ValueOf(v)
	rows := make([]table.Row, iv.Len())

	for i := 0; i < iv.Len(); i++ {
		row := make(table.Row, len(headers))
		ev := iv.Index(i)

		for j := range headers {
			f := ev.Field(j)

			if t, ok := f.Interface().(time.Time); ok {
				row[j] = t.Format(time.Stamp)
			} else if s, ok := f.Interface().([]string); ok {
				row[j] = strings.Join(s, ",")
			} else {
				row[j] = fmt.Sprint(f.Interface())
			}
		}

		rows[i] = row
	}

	return rows
}

func collectHeaders(t reflect.Type) (table.Row, []table.ColumnConfig) {
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

	return headers, cols
}
