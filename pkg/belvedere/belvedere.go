package belvedere

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/internal/gcp"
	"go.opencensus.io/trace"
	compute "google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/dns/v1"
	"google.golang.org/api/logging/v2"
	"google.golang.org/api/secretmanager/v1"
)

// Project provides the main functionality of Belvedere.
type Project interface {
	// Name returns the name of the project.
	Name() string

	// Setup enables all required GCP services, grants Deployment Manager the permissions required
	// to manage project accounts and IAM roles, and creates a deployment with the base resources
	// needed to use Belvedere.
	Setup(ctx context.Context, dnsZone string, dryRun bool, interval time.Duration) error

	// Teardown deletes the shared firewall rules and managed zone created by Setup. It does not
	// disable services or downgrade Deployment Manager's permissions.
	Teardown(ctx context.Context, dryRun, async bool, interval time.Duration) error

	// DNSServers returns a list of DNS servers which handle the project's managed zone.
	DNSServers(ctx context.Context) ([]DNSServer, error)

	// Instances returns a list of running instances in the project. If an app or release are
	// provided, limits the results to instances running the given app or release.
	Instances(ctx context.Context, app, release string) ([]Instance, error)

	// MachineTypes returns a list of GCE machine types which are available for the given project or
	// GCE region, if one is provided.
	MachineTypes(ctx context.Context, region string) ([]MachineType, error)

	// Logs returns a logs service.
	Logs() LogService

	// Secrets returns a secrets service.
	Secrets() SecretsService

	// Apps returns an apps service.
	Apps() AppService

	Releases() ReleaseService
}

// DNSServer is a DNS server run by Google.
type DNSServer struct {
	Server string
}

// NewProject returns a new Project instance for the given GCP project. If no project is provided,
// the active project configured for the Google Cloud SDK is used.
func NewProject(ctx context.Context, name string) (Project, error) {
	if name == "" {
		s, err := activeProject()
		if err != nil {
			return nil, err
		}
		name = s
	}

	ls, err := logging.NewService(ctx)
	if err != nil {
		return nil, err
	}

	sm, err := secretmanager.NewService(ctx)
	if err != nil {
		return nil, err
	}

	ds, err := dns.NewService(ctx)
	if err != nil {
		return nil, err
	}

	d := &dnsService{
		project: name,
		dns:     ds,
	}

	return &project{
		logs: &logService{
			project: name,
			clock:   time.Now,
			logs:    ls,
		},
		secrets: &secretsService{
			project: name,
			sm:      sm,
		},
		apps: &appService{
			project: name,
			dns:     d,
		},
		dns:  d,
		name: name,
	}, nil
}

type project struct {
	name     string
	logs     LogService
	secrets  SecretsService
	apps     *appService
	dns      *dnsService
	releases *releaseService
}

func (p *project) Apps() AppService {
	return p.apps
}

func (p *project) Secrets() SecretsService {
	return p.secrets
}

func (p *project) Logs() LogService {
	return p.logs
}

func (p *project) Releases() ReleaseService {
	return p.releases
}

func (p *project) Instances(ctx context.Context, app, release string) ([]Instance, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.project.Instances")
	span.AddAttributes(
		trace.StringAttribute("app", app),
		trace.StringAttribute("release", release),
	)
	defer span.End()

	// Get our GCE client.
	gce, err := gcp.Compute(ctx)
	if err != nil {
		return nil, err
	}

	var instances []Instance
	// List all instances.
	if err := gce.Instances.AggregatedList(p.name).Pages(ctx,
		func(list *compute.InstanceAggregatedList) error {
			for _, items := range list.Items {
				for _, inst := range items.Instances {
					// Filter by app and release. Limit to belvedere instances only.
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

	// Sort by name and return.
	sort.SliceStable(instances, func(i, j int) bool {
		return instances[i].Name < instances[j].Name
	})
	return instances, nil
}

func (p *project) Name() string {
	return p.name
}

func (p *project) DNSServers(ctx context.Context) ([]DNSServer, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.project.DNSServers")
	defer span.End()

	// Find the project'p managed zone.
	mz, err := p.dns.findManagedZone(ctx)
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

func activeProject() (string, error) {
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

// Instance is a Google Compute Engine VM instance.
type Instance struct {
	Name        string
	MachineType string `table:"Machine Type"`
	Zone        string
	Status      string
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

func (p *project) MachineTypes(ctx context.Context, region string) ([]MachineType, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.project.MachineTypes")
	span.AddAttributes(
		trace.StringAttribute("region", region),
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
	if err := gce.MachineTypes.AggregatedList(p.name).Pages(ctx,
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

	// Convert to our type.
	machineTypes := make([]MachineType, 0, len(mtMap))
	for _, v := range mtMap {
		machineTypes = append(machineTypes, MachineType{
			Name:   v.Name,
			CPU:    int(v.GuestCpus),
			Memory: Memory(v.MemoryMb),
		})
	}

	// Sort the machine types and return.
	sort.SliceStable(machineTypes, func(i, j int) bool {
		return machineTypes[i].lexical() < machineTypes[j].lexical()
	})
	return machineTypes, nil
}
