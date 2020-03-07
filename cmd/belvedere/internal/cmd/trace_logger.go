package cmd

import (
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/alessio/shellescape"
	"go.opencensus.io/trace"
)

type traceLogger struct {
	m sync.Mutex
}

func NewTraceLogger() trace.Exporter {
	return &traceLogger{}
}

func (l *traceLogger) ExportSpan(s *trace.SpanData) {
	l.m.Lock()
	defer l.m.Unlock()

	_, _ = fmt.Fprintf(os.Stderr, "%s: %s (%s)", s.EndTime.Format(time.Stamp), s.Name, s.SpanID)
	if s.Code != 0 {
		_, _ = fmt.Fprintf(os.Stderr, " code=%d", s.Code)
	}
	if s.Message != "" {
		_, _ = fmt.Fprintf(os.Stderr, " message=%s", shellescape.Quote(s.Message))
	}
	l.printAttributes(s.Attributes)
	_, _ = fmt.Fprintln(os.Stderr)

	for _, a := range s.Annotations {
		_, _ = fmt.Fprintf(os.Stderr, "  %s", shellescape.Quote(a.Message))
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
		v := attributes[k]
		if s, ok := v.(string); ok {
			v = shellescape.Quote(s)
		}
		_, _ = fmt.Fprintf(os.Stderr, " %v=%v", k, v)
	}
}
