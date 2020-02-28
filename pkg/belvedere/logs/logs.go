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

// Entry represents an app log entry.
type Entry struct {
	Timestamp time.Time
	Release   string
	Instance  string
	Container string
	Message   string
}

// Service manages access to application logs.
type Service interface {
	// List returns log entries from the given app which match the other optional criteria.
	List(ctx context.Context, app, release, instance string, maxAge time.Duration, filters []string) ([]Entry, error)
}

// NewService returns a new Service for the given project.
func NewService(ctx context.Context, project string) (Service, error) {
	ls, err := logging.NewService(ctx)
	if err != nil {
		return nil, err
	}
	return &service{
		project: project,
		ls:      ls,
		clock:   time.Now,
	}, nil
}

type service struct {
	project string
	ls      *logging.Service
	clock   func() time.Time
}

func (l *service) List(ctx context.Context, app, release, instance string, maxAge time.Duration, filters []string) ([]Entry, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.logs.list")
	span.AddAttributes(
		trace.StringAttribute("app", app),
		trace.Int64Attribute("max_age_ms", maxAge.Milliseconds()),
		trace.StringAttribute("app", app),
		trace.StringAttribute("release", release),
		trace.StringAttribute("instance", instance),
	)
	for i, f := range filters {
		span.AddAttributes(trace.StringAttribute(fmt.Sprintf("filter.%d", i), f))
	}
	defer span.End()

	filter := l.makeFilter(app, release, instance, maxAge, filters)

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
	return l.parse(entries)
}

func (l *service) makeFilter(app string, release string, instance string, maxAge time.Duration, filters []string) []string {
	// Always filter by resource type, time, and app.
	ts := l.clock().Add(-maxAge)
	filter := []string{
		`resource.type="gce_instance"`,
		fmt.Sprintf(`timestamp>=%q`, ts.Format(time.RFC3339Nano)),
		fmt.Sprintf(`jsonPayload.container.metadata.app=%q`, app),
	}

	// Include an optional release filter.
	if release != "" {
		filter = append(filter,
			fmt.Sprintf(`jsonPayload.container.metadata.release=%q`, release),
		)
	}

	// Include an optional instance filter.
	if instance != "" {
		filter = append(filter,
			fmt.Sprintf(`jsonPayload.instance.name=%q`, instance),
		)
	}

	// Include any other given filters.
	filter = append(filter, filters...)
	return filter
}

func (l *service) parse(entries *logging.ListLogEntriesResponse) ([]Entry, error) {
	results := make([]Entry, len(entries.Entries))
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
		results[i] = Entry{
			Timestamp: ts,
			Release:   payload.Container.Metadata.Release,
			Instance:  payload.Instance.Name,
			Container: strings.TrimPrefix(payload.Container.Name, "/"), // Docker prefixes this with a slash
			Message:   payload.Message,
		}
	}
	return results, nil
}
