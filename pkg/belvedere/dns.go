package belvedere

import (
	"context"
	"fmt"

	"go.opencensus.io/trace"
	"google.golang.org/api/dns/v1"
)

type dnsService struct {
	project string
	dns     *dns.Service
}

// findManagedZone returns the Cloud DNS managed zone created via Setup.
func (d *dnsService) findManagedZone(ctx context.Context) (*dns.ManagedZone, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.dns.findManagedZone")
	defer span.End()

	// Find the managed zone.
	mz, err := d.dns.ManagedZones.Get(d.project, "belvedere").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("error getting managed zone: %w", err)
	}
	return mz, nil
}
