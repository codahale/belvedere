package belvedere

import (
	"context"
	"testing"

	"github.com/codahale/belvedere/pkg/belvedere/internal/it"
	"github.com/google/go-cmp/cmp"
	compute "google.golang.org/api/compute/v0.beta"
	"gopkg.in/h2non/gock.v1"
)

func TestMachineTypes(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://www.googleapis.com/compute/beta/projects/my-project/aggregated/machineTypes?alt=json&prettyPrint=false").
		Reply(200).
		JSON(compute.MachineTypeAggregatedList{
			Items: map[string]compute.MachineTypesScopedList{
				"zones/us-central1-a": {
					MachineTypes: []*compute.MachineType{
						{
							Name:      "n1-standard-1",
							GuestCpus: 2,
							MemoryMb:  1024,
						},
					},
				},
				"zones/us-central1-b": {
					MachineTypes: []*compute.MachineType{
						{
							Name:      "n1-standard-1",
							GuestCpus: 2,
							MemoryMb:  1024,
						},
						{
							Name:      "n1-standard-4",
							GuestCpus: 4,
							MemoryMb:  4096,
						},
					},
				},
				"zones/us-west2-a": {
					MachineTypes: []*compute.MachineType{
						{
							Name:      "n1-standard-8",
							GuestCpus: 8,
							MemoryMb:  500,
						},
					},
				},
			},
		})

	actual, err := MachineTypes(context.TODO(), "my-project", "us-central1")
	if err != nil {
		t.Fatal(err)
	}

	expected := []MachineType{
		{
			Name:   "n1-standard-1",
			CPU:    2,
			Memory: 1024,
		},
		{
			Name:   "n1-standard-4",
			CPU:    4,
			Memory: 4096,
		},
	}

	if !cmp.Equal(expected, actual) {
		t.Error(cmp.Diff(expected, actual))
	}
}
