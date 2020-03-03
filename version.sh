#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

COMMIT="$(git describe --tags --long --always --dirty --broken)"
TIME="$(date -u +"%Y%m%d%H%M%S")"

cat << EOF > version.go
package main

//nolint:gochecknoinits
func init() {
  version = "${COMMIT}-${TIME}"
}
EOF
go fmt .
