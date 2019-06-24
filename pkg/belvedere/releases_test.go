package belvedere

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
)

func TestReleaseResources(t *testing.T) {
	config := &Config{
		Container: Container{
			Image:   "gcr.io/example/helloworld",
			Command: []string{"/usr/bin/helloworld"},
			Args:    []string{"one", "two"},
			Env: map[string]string{
				"ONE": "1",
				"TWO": "2",
			},
			DockerOptions: []string{"--turbo"},
		},
		Sidecars: map[string]Container{
			"nginx": {
				Image:   "gcr.io/example/nginx",
				Command: []string{"/usr/bin/nginx"},
				Args:    []string{"three", "four"},
				Env: map[string]string{
					"THREE": "3",
					"FOUR":  "4",
				},
				DockerOptions: []string{"--slow"},
			},
		},
		MinInstances:      10,
		MaxInstances:      100,
		InitialInstances:  20,
		UtilizationTarget: 0.6,
		MachineType:       "n1-standard-1",
	}
	resources := releaseResources("my-project", "us-central1", "my-app", "v43", "abcdef0123456789", config)

	have, err := json.MarshalIndent(map[string]interface{}{
		"resources": resources,
	}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	//_ = ioutil.WriteFile("release_fixture.json", have, 0644)

	want, err := ioutil.ReadFile("release_fixture.json")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(have, want) {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(string(want), string(have), false)
		if len(diffs) > 0 {
			t.Fatal(dmp.DiffPrettyText(diffs))
		}
	}
}
