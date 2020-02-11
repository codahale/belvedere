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
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/compute/v0.beta"
)

// WithInterval returns a new context with the given polling interval. This is required for using
// most methods in this package.
func WithInterval(ctx context.Context, interval time.Duration) context.Context {
	return waiter.WithInterval(ctx, interval)
}

// DNSServer is a DNS server run by Google.
type DNSServer struct {
	Server string
}

// DNSServers returns a list of DNS servers which handle the project's managed zone.
func DNSServers(ctx context.Context, project string) ([]DNSServer, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.DNSServers")
	span.AddAttributes(trace.StringAttribute("project", project))
	defer span.End()

	// Find the project's managed zone.
	mz, err := findManagedZone(ctx, project)
	if err != nil {
		return nil, err
	}

	// Return the DNS servers.
	var servers []DNSServer
	for _, s := range mz.NameServers {
		servers = append(servers, DNSServer{Server: s})
	}
	return servers, nil
}

// Instance is a Google Compute Engine VM instance.
type Instance struct {
	Name        string
	MachineType string `table:"Machine Type"`
	Zone        string
	Status      string
}

// Instances returns a list of running instances in the project. If an app or release are
// provided, limits the results to instances running the given app or release.
func Instances(ctx context.Context, project, app, release string) ([]Instance, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.Instances")
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
		return nil, fmt.Errorf("error listing zones: %w", err)
	}

	// Create an error group for the results.
	g, ctx := errgroup.WithContext(ctx)

	// Create a slice and mutex for aggregating results.
	var instances []Instance
	var m sync.Mutex

	// For each zone, start a goroutine to find instances.
	for _, zone := range zones.Items {
		g.Go(func() error {
			// Copy the zone name.
			zoneName := zone.Name

			// List all instances in the zone.
			zi, err := gce.Instances.List(project, zoneName).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("error listing instances in %s: %w", zoneName, err)
			}

			// Filter instances by app and release. Only return belvedere-managed instances,
			// regardless of criteria.
			for _, i := range zi.Items {
				if s, ok := i.Labels["belvedere-app"]; ok && (s == app || app == "") {
					if s, ok := i.Labels["belvedere-release"]; ok && (s == release || release == "") {
						// Aggregate instance names.
						m.Lock()
						mt := i.MachineType
						mt = mt[strings.LastIndex(mt, "/")+1:]
						zone := i.Zone
						zone = zone[strings.LastIndex(zone, "/")+1:]
						instances = append(instances, Instance{
							Name:        i.Name,
							MachineType: mt,
							Zone:        zone,
							Status:      i.Status,
						})
						m.Unlock()
					}
				}
			}

			return nil
		})
	}

	// Wait for all zones to complete.
	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Return results.
	return instances, nil
}

// SSH returns a function which execs to a Google Cloud SDK gcloud process which tunnels an SSH
// connection over IAP to the given instance.
func SSH(ctx context.Context, project, instance string, args []string) (func() error, error) {
	_, span := trace.StartSpan(ctx, "belvedere.SSH")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("instance", instance),
	)
	defer span.End()

	// Find gcloud on the path.
	gcloud, err := exec.LookPath("gcloud")
	if err != nil {
		return nil, fmt.Errorf("error finding gcloud executable: %w", err)
	}
	span.AddAttributes(trace.StringAttribute("gcloud", gcloud))

	sshArgs := append([]string{gcloud,
		"beta", "compute", "ssh", instance, "--tunnel-through-iap", "--",
	}, args...)
	return func() error {
		// Exec to gcloud.
		return syscall.Exec(gcloud, sshArgs, os.Environ())
	}, nil
}

// Memory represents a specific amount of RAM provided to a virtual machine.
type Memory int64

func (m Memory) String() string {
	if m < 1024 {
		return fmt.Sprintf("%6d MiB", m)
	}

	if m < (1024 * 1024) {
		return fmt.Sprintf("%6.2f GiB", float64(m)/1024)
	}

	return fmt.Sprintf("%6.2f TiB", float64(m)/1024/1024)
}

var _ fmt.Stringer = Memory(0)

// MachineType is a GCE machine type which can run VMs.
type MachineType struct {
	Name   string
	CPU    int
	Memory Memory
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
// GCE region, if one is provided.
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

	// Aggregate across pages of results.
	mtMap := map[string]*compute.MachineType{}
	region = "zones/" + region

	// Iterate through all pages of the results.
	pageToken := ""
	for {
		list, err := gce.MachineTypes.AggregatedList(project).
			MaxResults(1000).
			PageToken(pageToken).
			Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("error getting machine types list: %w", err)
		}

		// Aggregate across zones.
		for zone, items := range list.Items {
			if strings.HasPrefix(zone, region) {
				for _, mt := range items.MachineTypes {
					mtMap[mt.Name] = mt
				}
			}
		}

		if list.NextPageToken == "" {
			break
		}
		pageToken = list.NextPageToken
	}

	// Convert to our type in a sortable structure.
	var machineTypes machineTypeSlice
	for _, v := range mtMap {
		machineTypes = append(machineTypes, MachineType{
			Name:   v.Name,
			CPU:    int(v.GuestCpus),
			Memory: Memory(v.MemoryMb),
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
