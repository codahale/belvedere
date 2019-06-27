package belvedere

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/internal/gcp"
	"github.com/codahale/belvedere/pkg/belvedere/internal/waiter"
	"go.opencensus.io/trace"
	"golang.org/x/xerrors"
	compute "google.golang.org/api/compute/v0.beta"
)

// WithInterval returns a new context with the given polling interval. This is required for using
// most methods in this package.
func WithInterval(ctx context.Context, interval time.Duration) context.Context {
	return waiter.WithInterval(ctx, interval)
}

// DNSServers returns a list of DNS servers which handle the project's managed zone.
func DNSServers(ctx context.Context, project string) ([]string, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.DNSServers")
	span.AddAttributes(trace.StringAttribute("project", project))
	defer span.End()

	// Find the project's managed zone.
	mz, err := findManagedZone(ctx, project)
	if err != nil {
		return nil, err
	}

	// Return the DNS servers.
	var dnsServers []string
	for _, s := range mz.NameServers {
		dnsServers = append(dnsServers, s)
	}
	return dnsServers, nil
}

// ListInstances returns a list of running instances in the project. If an app or release are
// provided, limits the results to instances running the given app or release.
func ListInstances(ctx context.Context, project, app, release string) ([]string, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.ListInstances")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", app),
		trace.StringAttribute("release", release),
	)
	defer span.End()

	// Get our GCE client.
	gce, err := gcp.Compute(ctx)
	if err != nil {
		return nil, err
	}

	// List all zones in the project.
	zones, err := gce.Zones.List(project).Context(ctx).Do()
	if err != nil {
		return nil, xerrors.Errorf("error listing zones: %w", err)
	}

	// Create a wait group for the zones.
	var wg sync.WaitGroup
	wg.Add(len(zones.Items))

	// Create a slice and mutex for aggregating results.
	var instances []string
	var m sync.Mutex

	// For each zone, start a goroutine to find instances.
	for _, zone := range zones.Items {
		go func(zoneName string) {
			defer wg.Done()

			// List all instances in the zone.
			zi, err := gce.Instances.List(project, zoneName).Context(ctx).Do()
			if err != nil {
				return // ignore errors b/c concurrency sucks
			}

			// Filter instances by app and release. Only return belvedere-managed instances,
			// regardless of criteria.
			for _, i := range zi.Items {
				if s, ok := i.Labels["belvedere-app"]; ok && (s == app || app == "") {
					if s, ok := i.Labels["belvedere-release"]; ok && (s == release || release == "") {
						// Aggregate instance names.
						m.Lock()
						instances = append(instances, i.Name)
						m.Unlock()
					}
				}
			}
		}(zone.Name)
	}

	// Wait for all zones to complete.
	wg.Wait()

	// Return results.
	return instances, nil
}

// SSH returns a function which execs to a Google Cloud SDK gcloud process which tunnels an SSH
// connection over IAP to the given instance.
func SSH(ctx context.Context, project, instance string) (func() error, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.SSH")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("instance", instance),
	)
	defer span.End()

	// Find gcloud on the path.
	gcloud, err := exec.LookPath("gcloud")
	if err != nil {
		return nil, xerrors.Errorf("error finding gcloud executable: %w", err)
	}
	span.AddAttributes(trace.StringAttribute("gcloud", gcloud))

	return func() error {
		// Exec to gcloud.
		args := []string{gcloud, "beta", "compute", "ssh", instance, "--tunnel-through-iap"}
		return syscall.Exec(gcloud, args, os.Environ())
	}, nil
}

// MachineType is a GCE machine type which can run VMs.
type MachineType struct {
	Name   string
	CPU    int
	Memory int
}

func (mt MachineType) lexical() string {
	parts := strings.SplitN(mt.Name, "-", 3)
	var n int
	if len(parts) > 2 {
		n, _ = strconv.Atoi(parts[2])
	}
	return fmt.Sprintf("%10s%10s%010d", parts[0], parts[1], n)
}

// MachineTypes returns a list of GCE machine types which are available for the given project or
// region, if one is provided.
func MachineTypes(ctx context.Context, project, region string) ([]MachineType, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.MachineTypes")
	span.AddAttributes(
		trace.StringAttribute("project", project),
	)
	defer span.End()

	if region != "" {
		span.AddAttributes(trace.StringAttribute("region", region))
	}

	gce, err := gcp.Compute(ctx)
	if err != nil {
		return nil, err
	}

	list, err := gce.MachineTypes.AggregatedList(project).Context(ctx).Do()
	if err != nil {
		return nil, xerrors.Errorf("error getting machine types list: %w", err)
	}

	// Aggregate across zones.
	mtMap := map[string]*compute.MachineType{}
	region = "zones/" + region
	for zone, items := range list.Items {
		if strings.HasPrefix(zone, region) {
			for _, mt := range items.MachineTypes {
				mtMap[mt.Name] = mt
			}
		}
	}

	// Convert to our type in a sortable structure.
	var machineTypes machineTypeSlice
	for _, v := range mtMap {
		machineTypes = append(machineTypes, MachineType{
			Name:   v.Name,
			CPU:    int(v.GuestCpus),
			Memory: int(v.MemoryMb),
		})
	}

	// Sort the machine types and return.
	sort.Stable(machineTypes)
	return machineTypes, nil
}

type machineTypeSlice []MachineType

func (m machineTypeSlice) Len() int {
	return len(m)
}

func (m machineTypeSlice) Less(i, j int) bool {
	return m[i].lexical() < m[j].lexical()
}

func (m machineTypeSlice) Swap(i, j int) {
	tmp := m[i]
	m[i] = m[j]
	m[j] = tmp
}

var (
	rfc1035 = regexp.MustCompile(`^[[:alnum:]][[:alnum:]\-]{0,61}[[:alnum:]]|[[:alpha:]]$`)
)

// validateRFC1035 returns an error if the given name is not a valid RFC1305 DNS name.
func validateRFC1035(name string) error {
	if !rfc1035.MatchString(name) {
		return fmt.Errorf("invalid name: %s", name)
	}
	return nil
}
