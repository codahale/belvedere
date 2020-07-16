package setup

import (
	"context"
	"fmt"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/internal/check"
	"github.com/codahale/belvedere/pkg/belvedere/internal/waiter"
	"go.opencensus.io/trace"
	"google.golang.org/api/serviceusage/v1"
)

// The full set of GCP services required for Belvedere to be a happy home.
//nolint:gochecknoglobals // can't have non-scalar consts
var requiredServices = []string{
	"cloudasset.googleapis.com",
	"cloudbuild.googleapis.com",
	"clouddebugger.googleapis.com",
	"clouderrorreporting.googleapis.com",
	"cloudkms.googleapis.com",
	"cloudprofiler.googleapis.com",
	"cloudresourcemanager.googleapis.com",
	"cloudtrace.googleapis.com",
	"compute.googleapis.com",
	"containeranalysis.googleapis.com",
	"containerregistry.googleapis.com",
	"containerscanning.googleapis.com",
	"deploymentmanager.googleapis.com",
	"dns.googleapis.com",
	"iam.googleapis.com",
	"iap.googleapis.com",
	"logging.googleapis.com",
	"monitoring.googleapis.com",
	"oslogin.googleapis.com",
	"secretmanager.googleapis.com",
	"stackdriver.googleapis.com",
	"storage-api.googleapis.com",
}

func (s *service) EnableAPIs(ctx context.Context, project string, dryRun bool, interval time.Duration) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.internal.setup.EnableAPIs")
	defer span.End()

	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.BoolAttribute("dry_run", dryRun),
	)

	// Divide the required services up into batches of at most 20 services.
	for _, serviceIDs := range batchStrings(requiredServices, 20) {
		if dryRun {
			fmt.Printf("Enable %v\n", serviceIDs)
			continue
		}

		// Enable the services.
		op, err := s.su.Services.BatchEnable(
			fmt.Sprintf("projects/%s", project),
			&serviceusage.BatchEnableServicesRequest{
				ServiceIds: serviceIDs,
			},
		).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("error batch enabling: %w", err)
		}

		// Record which services we enabled.
		for _, service := range serviceIDs {
			span.Annotate([]trace.Attribute{
				trace.StringAttribute("service", service),
				trace.StringAttribute("operation", op.Name),
			}, "Service enabled")
		}

		// Poll for the services to be enabled.
		if err := waiter.Poll(ctx, interval, check.SU(ctx, s.su, op.Name)); err != nil {
			return err
		}
	}

	return nil
}

func batchStrings(s []string, n int) [][]string {
	var b [][]string
	for n < len(s) {
		s, b = s[n:], append(b, s[0:n:n])
	}

	b = append(b, s)

	return b
}
