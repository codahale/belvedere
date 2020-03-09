package sshcmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	"github.com/peterbourgon/ff/v3/ffcli"
)

func New(root *rootcmd.Config) *ffcli.Command {
	fs := flag.NewFlagSet("belvedere ssh", flag.ExitOnError)
	root.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "ssh",
		ShortUsage: "belvedere ssh <instance> [<ssh-arg>...]",
		ShortHelp:  "SSH to an instance.",
		LongHelp: cmd.Wrap(`SSH to an instance.

This uses the Google Cloud SDK (gcloud) to open an SSH connection to the specified instance. It does
so using an IAP tunnel, allowing Belvedere to block all public access to the SSH port on your GCE
instances. The only SSH traffic allowed in is via the IAP tunnel, which ensures that all SSH access
requires GCP credentials and access to your GCP project. For more information on IAP tunneling, see
https://cloud.google.com/iap/docs/tcp-forwarding-overview.`),
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) < 1 {
				return flag.ErrHelp
			}
			instance := args[0]

			// Find gcloud on the path.
			gcloud, err := exec.LookPath("gcloud")
			if err != nil {
				return fmt.Errorf("error finding gcloud executable: %w", err)
			}

			// Concat SSH arguments.
			sshArgs := append([]string{
				gcloud, "compute", "ssh", instance, "--tunnel-through-iap", "--",
			}, args[1:]...)

			// Exec to gcloud as last bit of main.
			root.Callback = func() error {
				return syscall.Exec(gcloud, sshArgs, os.Environ())
			}
			return nil
		},
	}
}
