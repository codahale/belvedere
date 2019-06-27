package belvedere

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestContainer_DockerArgs(t *testing.T) {
	container := Container{
		Image:   "gcr.io/example/example",
		Command: "/usr/bin/example",
		Args:    []string{"-h", "-y"},
		Env: map[string]string{
			"TWO": "2",
			"ONE": "1",
		},
		DockerOptions: []string{"--privileged"},
	}

	expected := []string{
		"--log-driver", "gcplogs",
		"--log-opt", "labels=alphabet,env",
		"--name", "my-example",
		"--network", "host",
		"--oom-kill-disable",
		"--label", "alphabet=latin",
		"--label", "env=qa",
		"--env", "RELEASE=v3",
		"--env", "ONE=1",
		"--env", "TWO=2",
		"--privileged",
		"gcr.io/example/example@sha256:123456",
		"/usr/bin/example", "-h", "-y",
	}
	actual := dockerArgs(&container, "my-example", "v3", "123456",
		map[string]string{
			"env":      "qa",
			"alphabet": "latin",
		})

	if !cmp.Equal(expected, actual) {
		t.Fatal(cmp.Diff(expected, actual))
	}
}
