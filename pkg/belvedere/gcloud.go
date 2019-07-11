package belvedere

import (
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"time"

	"go.opencensus.io/trace"
	"golang.org/x/xerrors"
)

// DefaultProject returns the default project, if any, which the Google Cloud SDK is configured to
// use.
func DefaultProject(ctx context.Context) (string, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.DefaultProject")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "gcloud", "config", "config-helper", "--format=json")
	b, err := cmd.Output()
	if err != nil {
		return "", xerrors.Errorf("unable to execute gcloud: %w", err)
	}

	var config struct {
		Configuration struct {
			Properties struct {
				Core struct {
					Project string `json:"project"`
				} `json:"core"`
			} `json:"properties"`
		} `json:"configuration"`
	}
	if err = json.Unmarshal(b, &config); err != nil {
		return "", xerrors.Errorf("unable to parse config-helper output: %w", err)
	}

	p := config.Configuration.Properties.Core.Project
	if p != "" {
		span.AddAttributes(trace.StringAttribute("project", p))
		return p, nil
	}

	return "", errors.New("project not found")
}
