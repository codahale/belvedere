package belvedere

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/internal/gcp"
	"go.opencensus.io/trace"
	"golang.org/x/xerrors"
	"google.golang.org/api/logging/v2"
)

// LogEntry represents an app log entry.
type LogEntry struct {
	Timestamp time.Time
	Release   string
	Instance  string
	Container string
	Message   string
}

// Logs returns log entries from the given app which match the other optional criteria. minTimestamp
// is required.
func Logs(ctx context.Context, project, app, release, instance string, minTimestamp time.Time, filters []string) ([]LogEntry, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.Logs")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", app),
		trace.StringAttribute("min_timestamp", minTimestamp.Format(time.RFC3339)),
	)
	defer span.End()

	for i, f := range filters {
		span.AddAttributes(trace.StringAttribute(fmt.Sprintf("filter.%d", i), f))
	}

	// Get our Logging client.
	logs, err := gcp.Logging(ctx)
	if err != nil {
		return nil, err
	}

	// Always filter by resource type, time, and app.
	filter := []string{
		`resource.type="gce_instance"`,
		fmt.Sprintf(`timestamp>=%q`, minTimestamp.Format(time.RFC3339Nano)),
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
	entries, err := logs.Entries.List(&logging.ListLogEntriesRequest{
		Filter:        strings.Join(filter, " "),
		OrderBy:       "timestamp desc", // reverse chronological order
		ResourceNames: []string{fmt.Sprintf("projects/%s", project)},
		PageSize:      1000, // cap at 1000 entries
	}).Context(ctx).Do()
	if err != nil {
		return nil, xerrors.Errorf("error listing log entries: %w", err)
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
			return nil, xerrors.Errorf("error parsing log entry %s: %w", e.InsertId, err)
		}
		ts, err := time.Parse(time.RFC3339Nano, e.Timestamp)
		if err != nil {
			return nil, xerrors.Errorf("error parsing timestamp in %s: %w", e.InsertId, err)
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
