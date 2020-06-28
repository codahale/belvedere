package belvedere

import (
	"context"
	"net/http"
	"testing"

	"github.com/codahale/belvedere/internal/assert"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	"gopkg.in/h2non/gock.v1"
)

func TestProject_MachineTypes(t *testing.T) {
	defer gock.Off()

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/" +
		"aggregated/machineTypes?alt=json&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.MachineTypeAggregatedList{
			NextPageToken: "",
			Items: map[string]compute.MachineTypesScopedList{
				"zones/us-central1-a": {
					MachineTypes: []*compute.MachineType{
						{
							Name:        "n1-standard-1",
							GuestCpus:   2,
							MemoryMb:    1024,
							IsSharedCpu: false,
						},
					},
				},
				"zones/us-central1-b": {
					MachineTypes: []*compute.MachineType{
						{
							Name:        "n1-standard-1",
							GuestCpus:   2,
							MemoryMb:    1024,
							IsSharedCpu: false,
						},
						{
							Name:        "n1-standard-4",
							GuestCpus:   4,
							MemoryMb:    4096,
							IsSharedCpu: true,
						},
					},
				},
				"zones/us-west2-a": {
					MachineTypes: []*compute.MachineType{
						{
							Name:        "n1-standard-8",
							GuestCpus:   8,
							MemoryMb:    500,
							IsSharedCpu: false,
						},
					},
				},
			},
		})

	s, err := NewProject(
		context.Background(),
		"my-project",
		option.WithHTTPClient(http.DefaultClient),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	got, err := s.MachineTypes(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}

	want := []MachineType{
		{
			Name:      "n1-standard-1",
			CPU:       2,
			Memory:    1024,
			SharedCPU: false,
		},
		{
			Name:      "n1-standard-4",
			CPU:       4,
			Memory:    4096,
			SharedCPU: true,
		},
		{
			Name:      "n1-standard-8",
			CPU:       8,
			Memory:    500,
			SharedCPU: false,
		},
	}

	assert.Equal(t, "MachineTypes()", want, got)
}

func TestProject_MachineTypes_Region(t *testing.T) {
	defer gock.Off()

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/" +
		"aggregated/machineTypes?alt=json&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.MachineTypeAggregatedList{
			NextPageToken: "",
			Items: map[string]compute.MachineTypesScopedList{
				"zones/us-central1-a": {
					MachineTypes: []*compute.MachineType{
						{
							Name:        "n1-standard-1",
							GuestCpus:   2,
							MemoryMb:    1024,
							IsSharedCpu: false,
						},
					},
				},
				"zones/us-central12-b": {
					MachineTypes: []*compute.MachineType{
						{
							Name:        "n1-standard-1",
							GuestCpus:   2,
							MemoryMb:    1024,
							IsSharedCpu: false,
						},
						{
							Name:        "n1-standard-4",
							GuestCpus:   4,
							MemoryMb:    4096,
							IsSharedCpu: false,
						},
					},
				},
				"zones/us-west2-a": {
					MachineTypes: []*compute.MachineType{
						{
							Name:        "n1-standard-8",
							GuestCpus:   8,
							MemoryMb:    500,
							IsSharedCpu: false,
						},
					},
				},
			},
		})

	s, err := NewProject(
		context.Background(),
		"my-project",
		option.WithHTTPClient(http.DefaultClient),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	got, err := s.MachineTypes(context.Background(), "us-central1")
	if err != nil {
		t.Fatal(err)
	}

	want := []MachineType{
		{
			Name:      "n1-standard-1",
			CPU:       2,
			Memory:    1024,
			SharedCPU: false,
		},
	}

	assert.Equal(t, "MachineTypes()", want, got)
}
