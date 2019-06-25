package cloudinit

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCloudConfig(t *testing.T) {
	config := CloudConfig{
		Files: []File{
			{
				Path:        "/etc/password/",
				Content:     "one\ntwo\nthree\n",
				Permissions: "0666",
			},
			{
				Path:        "/etc/init",
				Content:     "four\nfive\nsix",
				Permissions: "0888",
			},
		},
		Commands: []string{
			"say moo moo moo moo",
			"reboot",
		},
	}

	actual := []byte(config.String())

	//_ = ioutil.WriteFile("cloudinit.yaml", actual, 0644)

	expected, err := ioutil.ReadFile("cloudinit.yaml")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(actual, expected) {
		t.Error(cmp.Diff(string(expected), string(actual)))
	}
}
