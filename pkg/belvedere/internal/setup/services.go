package setup

import (
	"context"
	"fmt"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/internal/check"
	"go.opencensus.io/trace"
	"google.golang.org/api/serviceusage/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	// The full set of GCP services required for Belvedere to be a happy home.
	requiredServices = []string{
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
		"stackdriver.googleapis.com",
		"storage-api.googleapis.com",
	}
)

// EnableServices enables all required services for the given GCP project.
func EnableServices(ctx context.Context, project string) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.internal.base.EnableServices")
	defer span.End()

	su, err := serviceusage.NewService(ctx)
	if err != nil {
		return err
	}

	// Divide the required services up into batches of at most 20 services.
	for _, serviceIDs := range batchStrings(requiredServices, 20) {
		// Enable the services.
		op, err := su.Services.BatchEnable(
			fmt.Sprintf("projects/%s", project),
			&serviceusage.BatchEnableServicesRequest{
				ServiceIds: serviceIDs,
			},
		).Do()
		if err != nil {
			return err
		}

		// Record which services we enabled.
		for _, service := range serviceIDs {
			span.Annotate([]trace.Attribute{
				trace.StringAttribute("service", service),
				trace.StringAttribute("operation", op.Name),
			}, "Service enabled")
		}

		// Wait for the services to be enabled.
		if err := wait.Poll(10*time.Second, 5*time.Minute, check.SU(ctx, su, op.Name)); err != nil {
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
