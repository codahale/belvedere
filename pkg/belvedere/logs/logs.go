package logs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.opencensus.io/trace"
	"google.golang.org/api/logging/v2"
)

// LogService manages access to application logs.
type LogService interface {
	// List returns log entries from the given app which match the other optional criteria.
	List(ctx context.Context, app, release, instance string, maxAge time.Duration, filters []string) ([]LogEntry, error)
}

// LogEntry represents an app log entry.
type LogEntry struct {
	Timestamp time.Time
	Release   string
	Instance  string
	Container string
	Message   string
}

// NewLogService returns a new LogService for the given project.
func NewLogService(ctx context.Context, project string) (LogService, error) {
	ls, err := logging.NewService(ctx)
	if err != nil {
		return nil, err
	}
	return &logService{
		project: project,
		ls:      ls,
		clock:   time.Now,
	}, nil
}

type logService struct {
	project string
	ls      *logging.Service
	clock   func() time.Time
}

func (l *logService) List(ctx context.Context, app, release, instance string, maxAge time.Duration, filters []string) ([]LogEntry, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.logs.list")
	span.AddAttributes(
		trace.StringAttribute("app", app),
		trace.Int64Attribute("max_age_ms", maxAge.Milliseconds()),
	)
	defer span.End()

	for i, f := range filters {
		span.AddAttributes(trace.StringAttribute(fmt.Sprintf("filter.%d", i), f))
	}

	// Always filter by resource type, time, and app.
	ts := l.clock().Add(-maxAge)
	filter := []string{
		`resource.type="gce_instance"`,
		fmt.Sprintf(`timestamp>=%q`, ts.Format(time.RFC3339Nano)),
		fmt.Sprintf(`jsonPayload.container.metadata.app=%q`, app),
	}

	// Include an optional release filter.
	if release != "" {
		span.AddAttributes(trace.StringAttribute("release", release))
		filter = append(filter,
			fmt.Sprintf(`jsonPayload.container.metadata.release=%q`, release),
		)
	}

	// Include an optional instance filter.
	if instance != "" {
		span.AddAttributes(trace.StringAttribute("instance", instance))
		filter = append(filter,
			fmt.Sprintf(`jsonPayload.instance.name=%q`, instance),
		)
	}

	// Include any other given filters.
	filter = append(filter, filters...)

	// List log entries which match the full set of filters.
	entries, err := l.ls.Entries.List(&logging.ListLogEntriesRequest{
		Filter:        strings.Join(filter, " "),
		OrderBy:       "timestamp desc", // reverse chronological order
		ResourceNames: []string{fmt.Sprintf("projects/%s", l.project)},
		PageSize:      1000, // cap at 1000 entries
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("error listing log entries: %w", err)
	}

	// Parse the resulting log entries to return structured data.
	results := make([]LogEntry, len(entries.Entries))
	var payload struct {
		Container struct {
			Name     string `json:"name"`
			Metadata struct {
				Release string `json:"release"`
			} `json:"metadata"`
		} `json:"container"`
		Instance struct {
			Name string `json:"name"`
		} `json:"instance"`
		Message string `json:"message"`
	}
	for i, e := range entries.Entries {
		if err := json.Unmarshal(e.JsonPayload, &payload); err != nil {
			return nil, fmt.Errorf("error parsing log entry %s: %w", e.InsertId, err)
		}
		ts, err := time.Parse(time.RFC3339Nano, e.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("error parsing timestamp in %s: %w", e.InsertId, err)
		}
		results[i] = LogEntry{
			Timestamp: ts,
			Release:   payload.Container.Metadata.Release,
			Instance:  payload.Instance.Name,
			Container: strings.TrimPrefix(payload.Container.Name, "/"), // Docker prefixes this with a slash
			Message:   payload.Message,
		}
	}
	return results, nil
}
