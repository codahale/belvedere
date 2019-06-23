package belvedere

import (
	"context"
	"io"
	"os"
	"sync"

	"go.opencensus.io/trace"
	"google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/dns/v1"
)

func DNSServers(ctx context.Context, project string) ([]string, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.DNSServers")
	span.AddAttributes(trace.StringAttribute("project", project))
	defer span.End()

	d, err := dns.NewService(ctx)
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

func ListInstances(ctx context.Context, project, appName, relName string) ([]string, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.ListInstances")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", appName),
		trace.StringAttribute("release", relName),
	)
	defer span.End()

	gce, err := compute.NewService(ctx)
	if err != nil {
		return nil, err
	}

	zones, err := gce.Zones.List(project).Do()
	if err != nil {
		return nil, err
	}

	c := make(chan string, 100)
	var wg sync.WaitGroup

	for _, zone := range zones.Items {
		wg.Add(1)
		go func(zoneName string) {
			defer wg.Done()
			zi, err := gce.Instances.List(project, zoneName).Do()
			if err != nil {
				return
			}

			for _, i := range zi.Items {
				if s, ok := i.Labels["belvedere-app"]; ok && (s == appName || appName == "") {
					if s, ok := i.Labels["belvedere-release"]; ok && (s == relName || relName == "") {
						c <- i.Name
					}
				}
			}
		}(zone.Name)
	}
	wg.Wait()
	close(c)

	var instances []string
	for s := range c {
		instances = append(instances, s)
	}

	return instances, nil
}

func openPath(path string) (io.ReadCloser, error) {
	if path == "-" {
		return os.Stdin, nil
	}
	return os.Open(path)
}
