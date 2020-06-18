package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cli"
	"github.com/spf13/cobra"
)

func newSSHCmd() *cli.Command {
	return &cli.Command{
		UI: cobra.Command{
			Use:     `ssh <instance> [<ssh-arg>...]`,
			Example: `ssh my-app-v43-hxht -- ls -al`,
			Short:   `SSH to an instance`,
			Long: `SSH to an instance.

This uses the Google Cloud SDK (gcloud) to open an SSH connection to the specified instance. It does
so using an IAP tunnel, allowing Belvedere to block all public access to the SSH port on your GCE
instances. The only SSH traffic allowed in is via the IAP tunnel, which ensures that all SSH access
requires GCP credentials and access to your GCP project. For more information on IAP tunneling, see
https://cloud.google.com/iap/docs/tcp-forwarding-overview.`,
			Args: cobra.MinimumNArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
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
				return syscall.Exec(gcloud, sshArgs, os.Environ())
			},
		},
	}
}
