package belvedere

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sync"
	"syscall"

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

func ListInstances(ctx context.Context, project, app, release string) ([]string, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.ListInstances")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", app),
		trace.StringAttribute("release", release),
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
				if s, ok := i.Labels["belvedere-app"]; ok && (s == app || app == "") {
					if s, ok := i.Labels["belvedere-release"]; ok && (s == release || release == "") {
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

func SSH(ctx context.Context, project, instance string) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.SSH")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("instance", instance),
	)
	gcloud, err := exec.LookPath("gcloud")
	if err != nil {
		return err
	}
	span.AddAttributes(trace.StringAttribute("gcloud", gcloud))
	span.End()

	return syscall.Exec(
		gcloud,
		[]string{gcloud, "beta", "compute", "ssh", instance, "--tunnel-through-iap"},
		os.Environ(),
	)
}

var (
	rfc1035 = regexp.MustCompile(`^[[:alnum:]][[:alnum:]\-]{0,61}[[:alnum:]]|[[:alpha:]]$`)
)

func validateRFC1035(name string) error {
	if !rfc1035.MatchString(name) {
		return fmt.Errorf("invalid name: %s", name)
	}
	return nil
}
