package gcp

import (
	"fmt"
	"regexp"
)

var (
	rfc1035 = regexp.MustCompile(`^[[:alnum:]][[:alnum:]\-]{0,61}[[:alnum:]]|[[:alpha:]]$`)
)

// ValidateRFC1035 returns an error if the given name is not a valid RFC1305 DNS name.
func ValidateRFC1035(name string) error {
	if !rfc1035.MatchString(name) {
		return fmt.Errorf("invalid name: %s", name)
	}
	return nil
}
