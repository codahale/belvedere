package belvedere

import (
	"fmt"
	"sort"
	"time"

	"go.opencensus.io/trace"
)

type traceLogger struct {
}

func NewTraceLogger() trace.Exporter {
	return &traceLogger{}
}

func (l *traceLogger) ExportSpan(s *trace.SpanData) {
	fmt.Printf("%s: %s", s.EndTime.Format(time.Stamp), s.Name)
	if s.Code != 0 {
		fmt.Printf(" code=%d", s.Code)
	}
	if s.Message != "" {
		fmt.Printf(" message=%s", s.Message)
	}
	l.printAttributes(s.Attributes)
	fmt.Println()

	for _, a := range s.Annotations {
		fmt.Printf("  %s", a.Message)
		l.printAttributes(a.Attributes)
		fmt.Println()
	}
}

func (l *traceLogger) printAttributes(attributes map[string]interface{}) {
	var keys []string
	for k := range attributes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf(" %v=%v", k, attributes[k])
	}
}
