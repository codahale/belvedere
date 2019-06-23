package cloudinit

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

type User struct {
	Name string `yaml:"name,omitempty"`
	UID  int    `yaml:"uid,omitempty"`
}

type File struct {
	Path        string `yaml:"path,omitempty"`
	Permissions string `yaml:"permissions,omitempty"`
	Owner       string `yaml:"owner,omitempty"`
	Content     string `yaml:"content,omitempty"`
}

type CloudConfig struct {
	Files    []File   `yaml:"write_files,omitempty"`
	Commands []string `yaml:"runcmd,omitempty"`
}

func (c *CloudConfig) String() string {
	y, _ := yaml.Marshal(c)
	return fmt.Sprintf("#cloud-configs\n%s", string(y))
}
