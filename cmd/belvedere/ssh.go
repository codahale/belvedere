package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cli"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func newSSHCmd() *cli.Command {
	return &cli.Command{
		UI: cobra.Command{
			Use:   "ssh <instance> [<ssh-arg>...]",
			Short: "SSH to an instance",
			Long: `SSH to an instance.

This uses the Google Cloud SDK (gcloud) to open an SSH connection to the specified instance. It does
so using an IAP tunnel, allowing Belvedere to block all public access to the SSH port on your GCE
instances. The only SSH traffic allowed in is via the IAP tunnel, which ensures that all SSH access
requires GCP credentials and access to your GCP project. For more information on IAP tunneling, see
https://cloud.google.com/iap/docs/tcp-forwarding-overview.`,
			Args: cobra.MinimumNArgs(1),
		},
		RunCallback: func(ctx context.Context, project belvedere.Project, output cli.Output, fs afero.Fs, args []string) (func() error, error) {
			instance := args[0]

			// Find gcloud on the path.
			gcloud, err := exec.LookPath("gcloud")
			if err != nil {
				return nil, fmt.Errorf("error finding gcloud executable: %w", err)
			}

			// Concat SSH arguments.
			sshArgs := append([]string{
				gcloud, "compute", "ssh", instance, "--tunnel-through-iap", "--",
			}, args[1:]...)

			// Exec to gcloud as last bit of main.
			return func() error {
				return syscall.Exec(gcloud, sshArgs, os.Environ())
			}, nil
		},
	}
}
