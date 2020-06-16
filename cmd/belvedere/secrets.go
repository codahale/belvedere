package main

import (
	"context"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cli"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func newSecretsCmd() *cli.Command {
	return &cli.Command{
		UI: cobra.Command{
			Use:   "secrets",
			Short: "Commands for managing secrets",
			Long: `Commands for managing secrets.

Secrets are stored in Google Secret Manager, which provides integrity and confidentiality both at
rest and in flight, strong audit logging, and access control via IAM permissions. Secrets' values
are versioned, allowing for update rollouts and rollbacks.
`,
		},
		Subcommands: []*cli.Command{
			newSecretsListCmd(),
			newSecretsCreateCmd(),
			newSecretsUpdateCmd(),
			newSecretsDeleteCmd(),
			newSecretsGrantCmd(),
			newSecretsRevokeCmd(),
		},
	}
}

func newSecretsListCmd() *cli.Command {
	return &cli.Command{
		UI: cobra.Command{
			Use:   "list",
			Short: "List secrets",
			Long: `List secrets.

Because applications may share secrets (e.g. two applications both need to use the same API key),
secrets exist in their own namespace.`,
		},
		Run: func(ctx context.Context, project belvedere.Project, output cli.Output, fs afero.Fs, args []string) error {
			apps, err := project.Secrets().List(ctx)
			if err != nil {
				return err
			}
			return output.Print(apps)
		},
	}
}

func newSecretsCreateCmd() *cli.Command {
	var mf cli.ModifyFlags
	return &cli.Command{
		UI: cobra.Command{
			Use:   "create <name> [<data-file>]",
			Short: "Create a secret",
			Long: `Create a secret.

Creates a new secret with a value that is the contents of data-file, read as a bytestring.

If data-file is not specified (or is specified as '-'), the secret's value is read from STDIN
instead.`,
			Args: cobra.RangeArgs(1, 2),
		},
		Flags: func(fs *pflag.FlagSet) {
			mf.Register(fs)
		},
		Run: func(ctx context.Context, project belvedere.Project, output cli.Output, fs afero.Fs, args []string) error {
			name := args[0]
			path := cli.PathFromArgs(args, 1)

			b, err := afero.ReadFile(fs, path)
			if err != nil {
				return err
			}

			return project.Secrets().Create(ctx, name, b, mf.DryRun)
		},
	}
}

func newSecretsUpdateCmd() *cli.Command {
	var mf cli.ModifyFlags
	return &cli.Command{
		UI: cobra.Command{
			Use:   "update <name> [<data-file>]",
			Short: "Update a secret",
			Long: `Update a secret.

Updates the secret's value to be the contents of data-file, read as a bytestring.

If data-file is not specified (or is specified as '-'), the secret's value is read from STDIN
instead.`,
			Args: cobra.RangeArgs(1, 2),
		},
		Flags: func(fs *pflag.FlagSet) {
			mf.Register(fs)
		},
		Run: func(ctx context.Context, project belvedere.Project, output cli.Output, fs afero.Fs, args []string) error {
			name := args[0]
			path := cli.PathFromArgs(args, 1)

			b, err := afero.ReadFile(fs, path)
			if err != nil {
				return err
			}

			return project.Secrets().Update(ctx, name, b, mf.DryRun)
		},
	}
}

func newSecretsDeleteCmd() *cli.Command {
	var mf cli.ModifyFlags
	return &cli.Command{
		UI: cobra.Command{
			Use:   "delete <name>",
			Short: "Delete a secret",
			Long: `Delete a secret.

This deletes all versions of the secret as well, and cannot be undone.`,
			Args: cobra.ExactArgs(1),
		},
		Flags: func(fs *pflag.FlagSet) {
			mf.Register(fs)
		},
		Run: func(ctx context.Context, project belvedere.Project, output cli.Output, fs afero.Fs, args []string) error {
			name := args[0]
			return project.Secrets().Delete(ctx, name, mf.DryRun)
		},
	}
}

func newSecretsGrantCmd() *cli.Command {
	var mf cli.ModifyFlags
	return &cli.Command{
		UI: cobra.Command{
			Use:   "grant <name> <app>",
			Short: "Grant an application access to a secret",
			Long: `Grant an application access to a secret.

This modifies the secret's IAM policy to allow the application's service account access to the
secrets' value.`,
			Args: cobra.ExactArgs(2),
		},
		Flags: func(fs *pflag.FlagSet) {
			mf.Register(fs)
		},
		Run: func(ctx context.Context, project belvedere.Project, output cli.Output, fs afero.Fs, args []string) error {
			name := args[0]
			app := args[1]
			return project.Secrets().Grant(ctx, name, app, mf.DryRun)
		},
	}
}

func newSecretsRevokeCmd() *cli.Command {
	var mf cli.ModifyFlags
	return &cli.Command{
		UI: cobra.Command{
			Use:   "revoke <name> <app>",
			Short: "Revoke an application's access to a secret",
			Long: `Revoke an application's access to a secret.

This modifies the secret's IAM policy to disallow the application's service account access to the
secrets' value.`,
			Args: cobra.ExactArgs(2),
		},
		Flags: func(fs *pflag.FlagSet) {
			mf.Register(fs)
		},
		Run: func(ctx context.Context, project belvedere.Project, output cli.Output, fs afero.Fs, args []string) error {
			name := args[0]
			app := args[1]
			return project.Secrets().Revoke(ctx, name, app, mf.DryRun)
		},
	}
}
