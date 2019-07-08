package main

//go:generate bash ./version.sh

import (
	"github.com/codahale/belvedere/cmd"
)

var (
	version = "unknown"
)

func main() {
	cmd.Execute(version)
}
