package dnsserverscmd

import (
	"context"
	"flag"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	"github.com/peterbourgon/ff/v2/ffcli"
)

type Config struct {
	root *rootcmd.Config
}

func New(root *rootcmd.Config) *ffcli.Command {
	config := Config{root: root}

	fs := flag.NewFlagSet("belvedere dns-servers", flag.ExitOnError)
	root.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "dns-servers",
		ShortUsage: "belvedere dns-servers [flags]",
		ShortHelp:  "List the project's DNS servers.",
		LongHelp: cmd.Wrap(`List the project's DNS servers.

These DNS servers should be registered in the domain's WHOIS record or otherwise have DNS requests
forwarded to them in order to resolve application hostnames to the load balancer IPs.`),
		FlagSet: fs,
		Exec:    config.Exec,
	}
}

func (c *Config) Exec(ctx context.Context, args []string) error {
	if len(args) != 0 {
		return flag.ErrHelp
	}
	servers, err := c.root.Project.DNSServers(ctx)
	if err != nil {
		return err
	}
	return c.root.Tables.Print(servers)
}
