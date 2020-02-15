package belvedere

import (
	"context"
	"testing"

	"github.com/codahale/belvedere/pkg/belvedere/internal/it"
	"github.com/google/go-cmp/cmp"
	compute "google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/dns/v1"
	"gopkg.in/h2non/gock.v1"
)

func TestDNSServers(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://dns.googleapis.com/dns/v1/projects/my-project/managedZones/belvedere?alt=json&prettyPrint=false").
		Reply(200).
		JSON(&dns.ManagedZone{
			NameServers: []string{"ns1.example.com", "ns2.example.com"},
		})

	actual, err := DNSServers(context.TODO(), "my-project")
	if err != nil {
		t.Fatal(err)
	}

	expected := []DNSServer{
		{Server: "ns1.example.com"},
		{Server: "ns2.example.com"},
	}

	if !cmp.Equal(expected, actual) {
		t.Fatal(cmp.Diff(expected, actual))
	}
}

func TestInstances(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://compute.googleapis.com/compute/beta/projects/my-project/aggregated/instances?alt=json&pageToken=&prettyPrint=false").
		Reply(200).
		JSON(&compute.InstanceAggregatedList{
			Items: map[string]compute.InstancesScopedList{
				"us-west-1a": {
					Instances: []*compute.Instance{
						{
							Name:        "non-belvedere-1",
							MachineType: "n1-standard-1",
							Status:      "RUNNING",
						},
						{
							Name:        "my-app-1",
							Zone:        "us-west-1a",
							MachineType: "n1-standard-1",
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
							Name:        "non-belvedere-2",
							MachineType: "n1-standard-1",
							Status:      "RUNNING",
						},
						{
							Name:        "my-app-2",
							Zone:        "us-west-1a",
							MachineType: "n1-standard-1",
							Status:      "RUNNING",
							Labels: map[string]string{
								"belvedere-app":     "my-app",
								"belvedere-release": "v2",
							},
						},
						{
							Name:        "another-app-1",
							Zone:        "us-west-1a",
							MachineType: "n1-standard-1",
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

	actual, err := Instances(context.TODO(), "my-project", "", "")
	if err != nil {
		t.Fatal(err)
	}

	expected := []Instance{
		{
			Name:        "another-app-1",
			MachineType: "n1-standard-1",
			Status:      "RUNNING",
			Zone:        "us-west-1a",
		},
		{
			Name:        "my-app-1",
			MachineType: "n1-standard-1",
			Status:      "RUNNING",
			Zone:        "us-west-1a",
		},
		{
			Name:        "my-app-2",
			MachineType: "n1-standard-1",
			Status:      "RUNNING",
			Zone:        "us-west-1a",
		},
	}

	if !cmp.Equal(expected, actual) {
		t.Fatal(cmp.Diff(expected, actual))
	}
}

func TestInstancesApp(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://compute.googleapis.com/compute/beta/projects/my-project/aggregated/instances?alt=json&pageToken=&prettyPrint=false").
		Reply(200).
		JSON(&compute.InstanceAggregatedList{
			Items: map[string]compute.InstancesScopedList{
				"us-west-1a": {
					Instances: []*compute.Instance{
						{
							Name:        "non-belvedere-1",
							MachineType: "n1-standard-1",
							Status:      "RUNNING",
						},
						{
							Name:        "my-app-1",
							Zone:        "us-west-1a",
							MachineType: "n1-standard-1",
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
							Name:        "non-belvedere-2",
							MachineType: "n1-standard-1",
							Status:      "RUNNING",
						},
						{
							Name:        "my-app-2",
							Zone:        "us-west-1a",
							MachineType: "n1-standard-1",
							Status:      "RUNNING",
							Labels: map[string]string{
								"belvedere-app":     "my-app",
								"belvedere-release": "v2",
							},
						},
						{
							Name:        "another-app-1",
							Zone:        "us-west-1a",
							MachineType: "n1-standard-1",
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

	actual, err := Instances(context.TODO(), "my-project", "my-app", "")
	if err != nil {
		t.Fatal(err)
	}

	expected := []Instance{
		{
			Name:        "my-app-1",
			MachineType: "n1-standard-1",
			Status:      "RUNNING",
			Zone:        "us-west-1a",
		},
		{
			Name:        "my-app-2",
			MachineType: "n1-standard-1",
			Status:      "RUNNING",
			Zone:        "us-west-1a",
		},
	}

	if !cmp.Equal(expected, actual) {
		t.Fatal(cmp.Diff(expected, actual))
	}
}

func TestInstancesAppRelease(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://compute.googleapis.com/compute/beta/projects/my-project/aggregated/instances?alt=json&pageToken=&prettyPrint=false").
		Reply(200).
		JSON(&compute.InstanceAggregatedList{
			Items: map[string]compute.InstancesScopedList{
				"us-west-1a": {
					Instances: []*compute.Instance{
						{
							Name:        "non-belvedere-1",
							MachineType: "n1-standard-1",
							Status:      "RUNNING",
						},
						{
							Name:        "my-app-1",
							Zone:        "us-west-1a",
							MachineType: "n1-standard-1",
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
							Name:        "non-belvedere-2",
							MachineType: "n1-standard-1",
							Status:      "RUNNING",
						},
						{
							Name:        "my-app-2",
							Zone:        "us-west-1a",
							MachineType: "n1-standard-1",
							Status:      "RUNNING",
							Labels: map[string]string{
								"belvedere-app":     "my-app",
								"belvedere-release": "v2",
							},
						},
						{
							Name:        "another-app-1",
							Zone:        "us-west-1a",
							MachineType: "n1-standard-1",
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

	actual, err := Instances(context.TODO(), "my-project", "my-app", "v2")
	if err != nil {
		t.Fatal(err)
	}

	expected := []Instance{
		{
			Name:        "my-app-2",
			MachineType: "n1-standard-1",
			Status:      "RUNNING",
			Zone:        "us-west-1a",
		},
	}

	if !cmp.Equal(expected, actual) {
		t.Fatal(cmp.Diff(expected, actual))
	}
}

func TestMachineTypes(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://compute.googleapis.com/compute/beta/projects/my-project/aggregated/machineTypes?alt=json&maxResults=1000&pageToken=&prettyPrint=false").
		Reply(200).
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
