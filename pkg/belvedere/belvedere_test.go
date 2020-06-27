package belvedere

import (
	"context"
	"net/http"
	"testing"

	"github.com/codahale/belvedere/internal/assert"
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/dns/v1"
	"google.golang.org/api/option"
	"gopkg.in/h2non/gock.v1"
)

func TestProject_DNSServers(t *testing.T) {
	defer gock.Off()

	gock.New("https://dns.googleapis.com/dns/v1/projects/my-project/managedZones/" +
		"belvedere?alt=json&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(&dns.ManagedZone{
			NameServers: []string{"ns1.example.com", "ns2.example.com"},
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

	got, err := s.DNSServers(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	want := []DNSServer{
		{Hostname: "ns1.example.com"},
		{Hostname: "ns2.example.com"},
	}

	assert.Equal(t, "DNSServers()", want, got)
}

func TestProject_Instances(t *testing.T) {
	defer gock.Off()

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/" +
		"aggregated/instances?alt=json&filter=labels.belvedere-app%21%3D%22%22&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(&compute.InstanceAggregatedList{
			Items: map[string]compute.InstancesScopedList{
				"us-west-1a": {
					Instances: []*compute.Instance{
						{
							Name:        "my-app-1",
							Zone:        "zones/us-west1-a",
							MachineType: "zones/us-west1-a/machineTypes/n1-standard-1",
							Status:      "RUNNING",
							Labels: map[string]string{
								"belvedere-app":     "my-app",
								"belvedere-release": "v1",
							},
						},
					},
				},
				"us-west-1b": {
					Instances: []*compute.Instance{
						{
							Name:        "my-app-2",
							Zone:        "zones/us-west1-a",
							MachineType: "zones/us-west1-a/machineTypes/n1-standard-1",
							Status:      "RUNNING",
							Labels: map[string]string{
								"belvedere-app":     "my-app",
								"belvedere-release": "v2",
							},
						},
						{
							Name:        "another-app-1",
							Zone:        "zones/us-west1-a",
							MachineType: "zones/us-west1-a/machineTypes/n1-standard-1",
							Status:      "RUNNING",
							Labels: map[string]string{
								"belvedere-app":     "another-app",
								"belvedere-release": "v1",
							},
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

	got, err := s.Instances(context.Background(), "", "")
	if err != nil {
		t.Fatal(err)
	}

	want := []Instance{
		{
			Name:        "another-app-1",
			MachineType: "n1-standard-1",
			Status:      "RUNNING",
			Zone:        "us-west1-a",
		},
		{
			Name:        "my-app-1",
			MachineType: "n1-standard-1",
			Status:      "RUNNING",
			Zone:        "us-west1-a",
		},
		{
			Name:        "my-app-2",
			MachineType: "n1-standard-1",
			Status:      "RUNNING",
			Zone:        "us-west1-a",
		},
	}

	assert.Equal(t, "Instances()", want, got)
}

func TestProject_InstancesApp(t *testing.T) {
	defer gock.Off()

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/" +
		"aggregated/instances?alt=json&" +
		"filter=labels.belvedere-app%21%3D%22%22+AND+labels.belvedere-app%3D%22my-app%22&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(&compute.InstanceAggregatedList{
			Items: map[string]compute.InstancesScopedList{
				"us-west-1a": {
					Instances: []*compute.Instance{
						{
							Name:        "my-app-1",
							Zone:        "zones/us-west1-a",
							MachineType: "zones/us-west1-a/machineTypes/n1-standard-1",
							Status:      "RUNNING",
							Labels: map[string]string{
								"belvedere-app":     "my-app",
								"belvedere-release": "v1",
							},
						},
					},
				},
				"us-west-1b": {
					Instances: []*compute.Instance{
						{
							Name:        "my-app-2",
							Zone:        "zones/us-west1-a",
							MachineType: "zones/us-west1-a/machineTypes/n1-standard-1",
							Status:      "RUNNING",
							Labels: map[string]string{
								"belvedere-app":     "my-app",
								"belvedere-release": "v2",
							},
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

	got, err := s.Instances(context.Background(), "my-app", "")
	if err != nil {
		t.Fatal(err)
	}

	want := []Instance{
		{
			Name:        "my-app-1",
			MachineType: "n1-standard-1",
			Status:      "RUNNING",
			Zone:        "us-west1-a",
		},
		{
			Name:        "my-app-2",
			MachineType: "n1-standard-1",
			Status:      "RUNNING",
			Zone:        "us-west1-a",
		},
	}

	assert.Equal(t, "Instances()", want, got)
}

func TestProject_InstancesAppRelease(t *testing.T) {
	defer gock.Off()

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/" +
		"aggregated/instances?alt=json&" +
		"filter=labels.belvedere-app%21%3D%22%22+AND+labels.belvedere-app%3D%22my-app%22+AND+" +
		"labels.belvedere-release%3D%22v2%22&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(&compute.InstanceAggregatedList{
			Items: map[string]compute.InstancesScopedList{
				"us-west-1b": {
					Instances: []*compute.Instance{
						{
							Name:        "my-app-2",
							Zone:        "zones/us-west1-a",
							MachineType: "zones/us-west1-a/machineTypes/n1-standard-1",
							Status:      "RUNNING",
							Labels: map[string]string{
								"belvedere-app":     "my-app",
								"belvedere-release": "v2",
							},
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

	got, err := s.Instances(context.Background(), "my-app", "v2")
	if err != nil {
		t.Fatal(err)
	}

	want := []Instance{
		{
			Name:        "my-app-2",
			MachineType: "n1-standard-1",
			Status:      "RUNNING",
			Zone:        "us-west1-a",
		},
	}

	assert.Equal(t, "Instances()", want, got)
}

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
							Name:      "n1-standard-8",
							GuestCpus: 8,
							MemoryMb:  500,
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
			Name:   "n1-standard-1",
			CPU:    2,
			Memory: 1024,
		},
		{
			Name:      "n1-standard-4",
			CPU:       4,
			Memory:    4096,
			SharedCPU: true,
		},
		{
			Name:   "n1-standard-8",
			CPU:    8,
			Memory: 500,
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
							Name:      "n1-standard-1",
							GuestCpus: 2,
							MemoryMb:  1024,
						},
					},
				},
				"zones/us-central12-b": {
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
			Name:   "n1-standard-1",
			CPU:    2,
			Memory: 1024,
		},
	}

	assert.Equal(t, "MachineTypes()", want, got)
}
