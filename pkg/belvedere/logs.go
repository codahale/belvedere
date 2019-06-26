package belvedere

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/internal/gcp"
	"go.opencensus.io/trace"
	"google.golang.org/api/logging/v2"
)

type Log struct {
	Timestamp time.Time
	Release   string
	Instance  string
	Container string
	Message   string
}

func Logs(ctx context.Context, project, app, release, instance string, minTimestamp time.Time, filters []string) ([]Log, error) {
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

	logs, err := gcp.Logging(ctx)
	if err != nil {
		return nil, err
	}

	filter := []string{
		fmt.Sprintf(`timestamp>=%q`, minTimestamp.Format(time.RFC3339Nano)),
		fmt.Sprintf(`jsonPayload.container.metadata.app=%q`, app),
	}

	if release != "" {
		span.AddAttributes(trace.StringAttribute("release", release))
		filter = append(filter,
			fmt.Sprintf(`jsonPayload.container.metadata.release=%q`, release),
		)
	}

	if instance != "" {
		span.AddAttributes(trace.StringAttribute("instance", instance))
		filter = append(filter,
			fmt.Sprintf(`jsonPayload.instance.name=%q`, instance),
		)
	}

	filter = append(filter, filters...)

	entries, err := logs.Entries.List(&logging.ListLogEntriesRequest{
		Filter:        strings.Join(filter, " "),
		OrderBy:       "timestamp desc",
		ResourceNames: []string{fmt.Sprintf("projects/%s", project)},
		PageSize:      1000,
	}).Do()
	if err != nil {
		return nil, err
	}

	results := make([]Log, len(entries.Entries))
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
			return nil, err
		}
		ts, err := time.Parse(time.RFC3339Nano, e.Timestamp)
		if err != nil {
			return nil, err
		}
		results[i] = Log{
			Timestamp: ts,
			Release:   payload.Container.Metadata.Release,
			Instance:  payload.Instance.Name,
			Container: strings.TrimPrefix(payload.Container.Name, "/"), // Docker prefixes this with a slash
			Message:   payload.Message,
		}
	}
	return results, nil
}
