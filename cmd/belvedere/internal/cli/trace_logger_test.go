package cli

import (
	"bytes"
	"testing"
	"time"

	"github.com/codahale/belvedere/internal/assert"
	"go.opencensus.io/trace"
)

func TestTraceLogger_ExportSpan(t *testing.T) {
	tests := []struct {
		name   string
		output string
		span   *trace.SpanData
	}{
		{
			name:   "simple",
			output: "Mar  8 17:26:00: example.thing.Func (0102030405060708)\n",
			span: &trace.SpanData{
				Name:    "example.thing.Func",
				EndTime: time.Date(2020, 3, 8, 17, 26, 0, 0, time.UTC),
				SpanContext: trace.SpanContext{
					SpanID: trace.SpanID{1, 2, 3, 4, 5, 6, 7, 8},
				},
			},
		},
		{
			name:   "message",
			output: "Mar  8 17:26:00: example.thing.Func (0102030405060708) message='it went ok'\n",
			span: &trace.SpanData{
				Name:    "example.thing.Func",
				EndTime: time.Date(2020, 3, 8, 17, 26, 0, 0, time.UTC),
				SpanContext: trace.SpanContext{
					SpanID: trace.SpanID{1, 2, 3, 4, 5, 6, 7, 8},
				},
				Status: trace.Status{
					Message: "it went ok",
				},
			},
		},
		{
			name:   "code",
			output: "Mar  8 17:26:00: example.thing.Func (0102030405060708) code=409\n",
			span: &trace.SpanData{
				Name:    "example.thing.Func",
				EndTime: time.Date(2020, 3, 8, 17, 26, 0, 0, time.UTC),
				SpanContext: trace.SpanContext{
					SpanID: trace.SpanID{1, 2, 3, 4, 5, 6, 7, 8},
				},
				Status: trace.Status{
					Code: 409,
				},
			},
		},
		{
			name:   "annotations",
			output: "Mar  8 17:26:00: example.thing.Func (0102030405060708)\n  Mar  7 17:26:00: 'it went ok' one='well yes ok'\n",
			span: &trace.SpanData{
				Name:    "example.thing.Func",
				EndTime: time.Date(2020, 3, 8, 17, 26, 0, 0, time.UTC),
				SpanContext: trace.SpanContext{
					SpanID: trace.SpanID{1, 2, 3, 4, 5, 6, 7, 8},
				},
				Annotations: []trace.Annotation{
					{
						Message: "it went ok",
						Time:    time.Date(2020, 3, 7, 17, 26, 0, 0, time.UTC),
						Attributes: map[string]interface{}{
							"one": "well yes ok",
						},
					},
				},
			},
		},
		{
			name:   "attributes",
			output: "Mar  8 17:26:00: example.thing.Func (0102030405060708) example='one two three' ok=200 other=yes\n",
			span: &trace.SpanData{
				Name:    "example.thing.Func",
				EndTime: time.Date(2020, 3, 8, 17, 26, 0, 0, time.UTC),
				SpanContext: trace.SpanContext{
					SpanID: trace.SpanID{1, 2, 3, 4, 5, 6, 7, 8},
				},
				Attributes: map[string]interface{}{
					"example": "one two three",
					"other":   "yes",
					"ok":      200,
				},
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			out := bytes.NewBuffer(nil)
			exporter := NewTraceLogger(out)
			exporter.ExportSpan(test.span)

			assert.Equal(t, "Output", test.output, out.String())
		})
	}
}
