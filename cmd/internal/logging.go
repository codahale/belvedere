package internal

import (
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"go.opencensus.io/trace"
)

type TraceLogger struct {
	m sync.Mutex
}

func (l *TraceLogger) ExportSpan(s *trace.SpanData) {
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

func (l *TraceLogger) printAttributes(attributes map[string]interface{}) {
	var keys []string
	for k := range attributes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		_, _ = fmt.Fprintf(os.Stderr, " %v=%v", k, attributes[k])
	}
}
