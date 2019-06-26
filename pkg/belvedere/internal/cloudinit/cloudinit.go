package cloudinit

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

type File struct {
	Path        string `yaml:"path,omitempty"`
	Permissions string `yaml:"permissions,omitempty"`
	Owner       string `yaml:"owner,omitempty"`
	Content     string `yaml:"content,omitempty"`
}

// A cloud-init YAML manifest.
// https://cloudinit.readthedocs.io/en/latest/topics/examples.html
type CloudConfig struct {
	WriteFiles  []File   `yaml:"write_files,omitempty"`
	RunCommands []string `yaml:"runcmd,omitempty"`
}

func (c *CloudConfig) String() string {
	y, _ := yaml.Marshal(c)
	return fmt.Sprintf("#cloud-config\n\n%s", string(y))
}
