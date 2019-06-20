package belvedere

import (
	"fmt"
	"sort"
	"time"

	"go.opencensus.io/trace"
)

type TraceLogger struct{}

func (TraceLogger) ExportSpan(s *trace.SpanData) {
	for _, a := range s.Annotations {
		fmt.Printf("%s: %s", a.Time.Format(time.Stamp), a.Message)
		var keys []string
		for k := range a.Attributes {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Printf(" %v=%v", k, a.Attributes[k])
		}
		fmt.Println()
	}
}

var _ trace.Exporter = &TraceLogger{}
