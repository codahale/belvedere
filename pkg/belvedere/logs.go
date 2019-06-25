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

func Logs(ctx context.Context, project, app, release, instance string, freshness time.Duration, filters []string) ([]string, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.Logs")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", app),
	)
	defer span.End()

	if instance != "" {
	}

	ctx, logs, err := gcp.Logging(ctx)
	if err != nil {
		return nil, err
	}

	filter := []string{
		fmt.Sprintf(`timestamp>=%q`, time.Now().Add(-freshness).Format(time.RFC3339Nano)),
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

	results := make([]string, len(entries.Entries))
	var payload struct {
		Container struct {
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
		results[i] = fmt.Sprintf("%s (%s/%s) %s",
			e.Timestamp, payload.Container.Metadata.Release, payload.Instance.Name, payload.Message)
	}
	return results, nil
}
