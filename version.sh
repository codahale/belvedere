#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

COMMIT="$(git rev-parse --short HEAD)"
TIME="$(date -u +"%Y%m%d%H%M%S")"

cat << EOF > version.go
package main

func init() {
  version = "${TIME}-${COMMIT}"
}
EOF
