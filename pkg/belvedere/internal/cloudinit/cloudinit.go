package cloudinit

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

// File represents a file to be created on instance boot.
type File struct {
	Path        string `yaml:"path,omitempty"`
	Permissions string `yaml:"permissions,omitempty"`
	Owner       string `yaml:"owner,omitempty"`
	Content     string `yaml:"content,omitempty"`
}

// CloudConfig contains a subset of cloud-init's cloud-config manifest properties.
// https://cloudinit.readthedocs.io/en/latest/topics/examples.html
type CloudConfig struct {
	WriteFiles  []File   `yaml:"write_files,omitempty"`
	RunCommands []string `yaml:"runcmd,omitempty"`
}

func (c *CloudConfig) String() string {
	y, _ := yaml.Marshal(c) // we explicitly use gopkg.in/yaml.v2 b/c it preserves ordering
	return fmt.Sprintf("#cloud-config\n\n%s", string(y))
}
