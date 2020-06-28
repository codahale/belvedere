package belvedere

import (
	"context"
	"fmt"
	"sort"
	"strconv"
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

func (mt MachineType) lexical() string {
	var n int

	parts := strings.SplitN(mt.Name, "-", 3)
	if len(parts) > 2 {
		n, _ = strconv.Atoi(parts[2])
	}

	return fmt.Sprintf("%10s%10s%010d", parts[0], parts[1], n)
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

	// Sort the machine types and return.
	sort.SliceStable(machineTypes, func(i, j int) bool {
		return machineTypes[i].lexical() < machineTypes[j].lexical()
	})

	return machineTypes
}
