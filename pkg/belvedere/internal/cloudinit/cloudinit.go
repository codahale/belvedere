package cloudinit

import (
	"encoding/json"
	"fmt"
)

// File represents a file to be created on instance boot.
type File struct {
	Path        string `json:"path,omitempty"`
	Permissions string `json:"permissions,omitempty"`
	Owner       string `json:"owner,omitempty"`
	Content     string `json:"content,omitempty"`
}

// CloudConfig contains a subset of cloud-init's cloud-config manifest properties.
// https://cloudinit.readthedocs.io/en/latest/topics/examples.html
type CloudConfig struct {
	WriteFiles  []File   `json:"write_files,omitempty"`
	RunCommands []string `json:"runcmd,omitempty"`
}

func (c *CloudConfig) String() string {
	j, _ := json.Marshal(c)
	return fmt.Sprintf("#cloud-config\n\n%s", j)
}
