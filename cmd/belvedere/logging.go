package main

import (
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"go.opencensus.io/trace"
)

type traceLogger struct {
	m sync.Mutex
}

var _ trace.Exporter = &traceLogger{}

func (l *traceLogger) ExportSpan(s *trace.SpanData) {
	l.m.Lock()
	defer l.m.Unlock()

	_, _ = fmt.Fprintf(os.Stderr, "%s: %s (%s)", s.EndTime.Format(time.Stamp), s.Name, s.SpanID)
	if s.Code != 0 {
		_, _ = fmt.Fprintf(os.Stderr, " code=%d", s.Code)
	}
	if s.Message != "" {
		_, _ = fmt.Fprintf(os.Stderr, " message=%s", s.Message)
	}
	l.printAttributes(s.Attributes)
	_, _ = fmt.Fprintln(os.Stderr)

	for _, a := range s.Annotations {
		_, _ = fmt.Fprintf(os.Stderr, "  %s", a.Message)
		l.printAttributes(a.Attributes)
		_, _ = fmt.Fprintln(os.Stderr)
	}
}

func (l *traceLogger) printAttributes(attributes map[string]interface{}) {
	keys := make([]string, 0, len(attributes))
	for k := range attributes {
		keys = append(keys, k)
	}
	sort.Stable(sort.StringSlice(keys))
	for _, k := range keys {
		_, _ = fmt.Fprintf(os.Stderr, " %v=%v", k, attributes[k])
	}
}
