package belvedere

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/codahale/belvedere/pkg/belvedere/internal/gcp"
	"go.opencensus.io/trace"
	compute "google.golang.org/api/compute/v0.beta"
)

// ActiveProject returns the project, if any, which the Google Cloud SDK is configured to use.
func ActiveProject(ctx context.Context) (string, error) {
	_, span := trace.StartSpan(ctx, "belvedere.ActiveProject")
	defer span.End()

	// Load SDK config.
	config, err := gcp.SDKConfig()
	if err != nil {
		return "", err
	}

	// Return core.project, if it exists.
	if core, ok := config["core"]; ok {
		if project, ok := core["project"]; ok {
			return project, nil
		}
	}

	// Complain if core.project doesn't exist.
	return "", fmt.Errorf("core.project not found")
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
	servers := make([]DNSServer, 0, len(mz.NameServers))
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

type instanceList []Instance

func (l instanceList) Len() int {
	return len(l)
}

func (l instanceList) Less(i, j int) bool {
	return l[i].Name < l[j].Name
}

func (l instanceList) Swap(i, j int) {
	tmp := l[i]
	l[i] = l[j]
	l[j] = tmp
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

	var instances instanceList
	// Filter by app and release. Limit to belvedere instances only.
	if err := gce.Instances.AggregatedList(project).Pages(ctx,
		func(list *compute.InstanceAggregatedList) error {
			for _, items := range list.Items {
				for _, inst := range items.Instances {
					if s, ok := inst.Labels["belvedere-app"]; ok && (s == app || app == "") {
						if s, ok := inst.Labels["belvedere-release"]; ok && (s == release || release == "") {
							mt := inst.MachineType
							mt = mt[strings.LastIndex(mt, "/")+1:]
							zone := inst.Zone
							zone = zone[strings.LastIndex(zone, "/")+1:]
							instances = append(instances, Instance{
								Name:        inst.Name,
								MachineType: mt,
								Zone:        zone,
								Status:      inst.Status,
							})
						}
					}
				}
			}
			return nil
		},
	); err != nil {
		return nil, fmt.Errorf("error listing instances: %w", err)
	}
	sort.Stable(instances)
	return instances, nil
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
	if err := gce.MachineTypes.AggregatedList(project).Pages(ctx,
		func(list *compute.MachineTypeAggregatedList) error {
			// Aggregate across zones.
			for zone, items := range list.Items {
				if strings.HasPrefix(zone, region) {
					for _, mt := range items.MachineTypes {
						mtMap[mt.Name] = mt
					}
				}
			}
			return nil
		},
	); err != nil {
		return nil, fmt.Errorf("error listing machine types: %w", err)
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
