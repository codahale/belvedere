// Package belvedere contains types and methods for deploying HTTP/2-based applications to Google
// Cloud Platform and managing them using best practices.
package belvedere

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/internal/check"
	"github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
	"github.com/codahale/belvedere/pkg/belvedere/internal/gcp"
	"github.com/codahale/belvedere/pkg/belvedere/internal/resources"
	"github.com/codahale/belvedere/pkg/belvedere/internal/setup"
	"go.opencensus.io/trace"
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/logging/v2"
	"google.golang.org/api/option"
	"google.golang.org/api/secretmanager/v1"
)

// Project provides the main functionality of Belvedere.
type Project interface {
	// Name returns the name of the project.
	Name() string

	// Setup enables all required GCP services, grants Deployment Service the permissions required
	// to manage project accounts and IAM roles, and creates a deployment with the base resources
	// needed to use Belvedere.
	Setup(ctx context.Context, dnsZone string, dryRun bool, interval time.Duration) error

	// Teardown deletes the shared firewall rules and managed zone created by Setup. It does not
	// disable services or downgrade Deployment Service's permissions.
	Teardown(ctx context.Context, dryRun, async bool, interval time.Duration) error

	// DNSServers returns a list of DNS servers which handle the project's managed zone.
	DNSServers(ctx context.Context) ([]DNSServer, error)

	// Instances returns a list of running instances in the project. If an application or release
	// are provided, limits the results to instances running the given app or release.
	Instances(ctx context.Context, app, release string) ([]Instance, error)

	// MachineTypes returns a list of GCE machine types which are available for the given project or
	// GCE region, if one is provided.
	MachineTypes(ctx context.Context, region string) ([]MachineType, error)

	// Logs provides methods for viewing application logs.
	Logs() LogService

	// Secrets provides methods for managing secrets.
	Secrets() SecretsService

	// Apps provides methods for managing applications.
	Apps() AppService

	// Releases provides methods for managing releases.
	Releases() ReleaseService
}

// DNSServer is a DNS server run by Google.
type DNSServer struct {
	Hostname string
}

// NewProject returns a new Project instance for the given GCP project.
func NewProject(ctx context.Context, name string, opts ...option.ClientOption) (Project, error) {
	if err := gcp.ValidateRFC1035(name); err != nil {
		return nil, err
	}

	ls, err := logging.NewService(ctx, opts...)
	if err != nil {
		return nil, err
	}

	sm, err := secretmanager.NewService(ctx, opts...)
	if err != nil {
		return nil, err
	}

	dm, err := deployments.NewManager(ctx, opts...)
	if err != nil {
		return nil, err
	}

	gce, err := compute.NewService(ctx, opts...)
	if err != nil {
		return nil, err
	}

	s, err := setup.NewService(ctx, opts...)
	if err != nil {
		return nil, err
	}

	res := resources.NewBuilder()
	hc := check.NewHealthChecker(gce)

	apps := &appService{
		project:   name,
		dm:        dm,
		setup:     s,
		gce:       gce,
		resources: res,
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
		apps: apps,
		releases: &releaseService{
			project:   name,
			dm:        dm,
			gce:       gce,
			resources: res,
			health:    hc,
			apps:      apps,
		},
		name:      name,
		dm:        dm,
		gce:       gce,
		setup:     s,
		resources: res,
	}, nil
}

type project struct {
	name      string
	logs      LogService
	secrets   SecretsService
	apps      *appService
	releases  *releaseService
	dm        deployments.Manager
	gce       *compute.Service
	setup     setup.Service
	resources resources.Builder
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
	defer span.End()

	span.AddAttributes(
		trace.StringAttribute("app", app),
		trace.StringAttribute("release", release),
	)

	// Construct filter set, always limiting results to Belvedere app instances.
	filters := []string{`labels.belvedere-app!=""`}
	if app != "" {
		filters = append(filters, fmt.Sprintf("labels.belvedere-app=%q", app))
	}

	if release != "" {
		filters = append(filters, fmt.Sprintf("labels.belvedere-release=%q", release))
	}

	var instances []Instance

	// List all instances.
	if err := p.gce.Instances.AggregatedList(p.name).
		Filter(strings.Join(filters, " AND ")).
		Pages(ctx,
			func(list *compute.InstanceAggregatedList) error {
				for _, items := range list.Items {
					for _, inst := range items.Instances {
						instances = append(instances, Instance{
							Name:        inst.Name,
							MachineType: lastPathComponent(inst.MachineType),
							Zone:        lastPathComponent(inst.Zone),
							Status:      inst.Status,
						})
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
	mz, err := p.setup.ManagedZone(ctx, p.name)
	if err != nil {
		return nil, err
	}

	// Return the DNS servers.
	servers := make([]DNSServer, 0, len(mz.NameServers))
	for _, s := range mz.NameServers {
		servers = append(servers, DNSServer{Hostname: s})
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

// Memory represents a specific amount of RAM provided to a virtual machine.
type Memory int64

func (m Memory) String() string {
	return fmt.Sprintf("%.2f", float64(m)/1024)
}

var _ fmt.Stringer = Memory(0)

// MachineType is a GCE machine type which can run VMs.
type MachineType struct {
	Name      string
	CPU       int    `table:"CPU,ralign"`
	Memory    Memory `table:"Memory (GiB),ralign"`
	SharedCPU bool   `table:"Shared CPU"`
}

func (mt MachineType) lexical() string {
	var n int

	parts := strings.SplitN(mt.Name, "-", 3)
	if len(parts) > 2 {
		n, _ = strconv.Atoi(parts[2])
	}

	return fmt.Sprintf("%10s%10s%010d", parts[0], parts[1], n)
}

//nolint:gocognit // this is just complicated
func (p *project) MachineTypes(ctx context.Context, region string) ([]MachineType, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.project.MachineTypes")
	defer span.End()

	span.AddAttributes(
		trace.StringAttribute("region", region),
	)

	// Limit by zone prefix.
	zonePrefix := "zones/"
	if region != "" {
		zonePrefix = zonePrefix + region + "-"
	}

	// Aggregate across pages of results.
	machineTypesByName := map[string]*compute.MachineType{}

	// Iterate through all pages of the results.
	if err := p.gce.MachineTypes.AggregatedList(p.name).Pages(ctx,
		func(list *compute.MachineTypeAggregatedList) error {
			for zone, items := range list.Items {
				if !strings.HasPrefix(zone, zonePrefix) {
					continue
				}

				for _, mt := range items.MachineTypes {
					machineTypesByName[mt.Name] = mt
				}
			}
			return nil
		},
	); err != nil {
		return nil, fmt.Errorf("error listing machine types: %w", err)
	}

	// Convert to our type.
	machineTypes := make([]MachineType, 0, len(machineTypesByName))
	for _, v := range machineTypesByName {
		machineTypes = append(machineTypes, MachineType{
			Name:      v.Name,
			CPU:       int(v.GuestCpus),
			Memory:    Memory(v.MemoryMb),
			SharedCPU: v.IsSharedCpu,
		})
	}

	// Sort the machine types and return.
	sort.SliceStable(machineTypes, func(i, j int) bool {
		return machineTypes[i].lexical() < machineTypes[j].lexical()
	})

	return machineTypes, nil
}

func lastPathComponent(s string) string {
	idx := strings.LastIndex(s, "/")
	if idx < 0 {
		return s
	}

	return s[idx+1:]
}
