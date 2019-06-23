package belvedere

import (
	"context"
	"io"
	"os"

	"go.opencensus.io/trace"
	dns2 "google.golang.org/api/dns/v1"
)

func DNSServers(ctx context.Context, project string) ([]string, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.DNSServers")
	span.AddAttributes(trace.StringAttribute("project", project))
	defer span.End()

	d, err := dns2.NewService(ctx)
	if err != nil {
		return nil, err
	}

	mz, err := d.ManagedZones.Get(project, "belvedere").Do()
	if err != nil {
		return nil, err
	}

	var dnsServers []string
	for _, s := range mz.NameServers {
		dnsServers = append(dnsServers, s)
	}
	return dnsServers, nil
}

func openPath(path string) (io.ReadCloser, error) {
	if path == "-" {
		return os.Stdin, nil
	}
	return os.Open(path)
}
