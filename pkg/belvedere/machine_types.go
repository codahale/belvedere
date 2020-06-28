package belvedere

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"go.opencensus.io/trace"
	"google.golang.org/api/compute/v1"
)

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

func (mt MachineType) lt(v MachineType) bool {
	a := strings.Split(mt.Name, "-")
	b := strings.Split(v.Name, "-")

	// Compare n1 vs e2, then standard vs highmem, then vCPU count.
	switch {
	case a[0] < b[0]:
		return true
	case a[0] > b[0]:
		return false
	case a[1] < b[1]:
		return true
	case a[1] > b[1]:
		return false
	default:
		return mt.CPU < v.CPU
	}
}

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
				// Skip zones outside the given region.
				if !strings.HasPrefix(zone, zonePrefix) {
					continue
				}

				// Aggregate machine types by name.
				for _, mt := range items.MachineTypes {
					machineTypesByName[mt.Name] = mt
				}
			}

			return nil
		},
	); err != nil {
		return nil, fmt.Errorf("error listing machine types: %w", err)
	}

	// Convert to our type, sort, and return.
	return machineTypesToSlice(machineTypesByName), nil
}

func machineTypesToSlice(machineTypesByName map[string]*compute.MachineType) []MachineType {
	machineTypes := make([]MachineType, 0, len(machineTypesByName))
	for _, v := range machineTypesByName {
		machineTypes = append(machineTypes, MachineType{
			Name:      v.Name,
			CPU:       int(v.GuestCpus),
			Memory:    Memory(v.MemoryMb),
			SharedCPU: v.IsSharedCpu,
		})
	}

	sort.SliceStable(machineTypes, func(i, j int) bool {
		return machineTypes[i].lt(machineTypes[j])
	})

	return machineTypes
}
