package gcp

import (
	"fmt"
	"regexp"
)

var (
	//nolint:gochecknoglobals // only compile once
	rfc1035 = regexp.MustCompile(`^[a-z]([-a-z0-9]*[a-z0-9])?$`)
)

type invalidNameError struct {
	Name string
}

func (e *invalidNameError) Error() string {
	return fmt.Sprintf("invalid name: %q", e.Name)
}

// ValidateRFC1035 returns an error if the given name is not a valid RFC1035 DNS name.
func ValidateRFC1035(name string) error {
	if len(name) < 1 || len(name) > 63 || !rfc1035.MatchString(name) {
		return &invalidNameError{Name: name}
	}

	return nil
}
