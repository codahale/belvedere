package cloudinit

import (
	"testing"

	"github.com/codahale/belvedere/pkg/belvedere/internal/fixtures"
)

func TestCloudConfig(t *testing.T) {
	config := CloudConfig{
		WriteFiles: []File{
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
		RunCommands: []string{
			"say moo moo moo moo",
			"reboot",
		},
	}

	fixtures.Compare(t, "cloudinit_fixture.yaml", []byte(config.String()))
}
