package main

import (
	"context"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cli"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/spf13/cobra"
)

func newDNSServersCmd() *cli.Command {
	return &cli.Command{
		UI: cobra.Command{
			Use:     `dns-servers`,
			Example: `belvedere dns-servers`,
			Short:   `List the project's DNS servers`,
			Long: `List the project's DNS servers.

These DNS servers should be registered in the domain's WHOIS record or otherwise have DNS requests
forwarded to them in order to resolve application hostnames to the load balancer IPs.`,
			Args: cobra.NoArgs,
		},
		Run: func(ctx context.Context, project belvedere.Project, in cli.Input, out cli.Output, args []string) error {
			servers, err := project.DNSServers(ctx)
			if err != nil {
				return err
			}
			return out.Print(servers)
		},
	}
}
