package deployments

import "fmt"

func Ref(name, property string) string {
	return fmt.Sprintf("$(ref.%s.%s)", name, property)
}

func SelfLink(name string) string {
	return Ref(name, "selfLink")
}

type Metadata struct {
	DependsOn []string `json:"dependsOn,omitempty"`
}

type Output struct {
	Name  string `json:"name"`
	Value string `json:"name"`
}

type Resource struct {
	Name       string      `json:"name"`
	Type       string      `json:"type"`
	Properties interface{} `json:"properties"`
	Metadata   *Metadata   `json:"metadata,omitempty"`
	Outputs    []Output    `json:"outputs,omitempty"`
}

type Config struct {
	Resources []Resource `json:"resources,omitempty"`
}
