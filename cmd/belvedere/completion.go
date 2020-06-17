package main

import (
	"github.com/codahale/belvedere/cmd/belvedere/internal/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func newCompletionCmd() *cli.Command {
	return &cli.Command{
		UI: cobra.Command{
			Use:    "completion",
			Hidden: true,
		},
		Subcommands: []*cli.Command{
			newCompletionBashCmd(),
			newCompletionFishCmd(),
			newCompletionPowershellCmd(),
			newCompletionZshCmd(),
		},
	}
}

func newCompletionBashCmd() *cli.Command {
	return &cli.Command{
		UI: cobra.Command{
			Use: "bash",
			RunE: func(cmd *cobra.Command, args []string) error {
				return cmd.Root().GenBashCompletion(cmd.OutOrStdout())
			},
		},
	}
}

func newCompletionFishCmd() *cli.Command {
	var includeDesc bool
	return &cli.Command{
		UI: cobra.Command{
			Use: "fish",
			RunE: func(cmd *cobra.Command, args []string) error {
				return cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
			},
		},
		Flags: func(fs *pflag.FlagSet) {
			fs.BoolVar(&includeDesc, "descriptions", false, "include descriptions")
		},
	}
}

func newCompletionPowershellCmd() *cli.Command {
	return &cli.Command{
		UI: cobra.Command{
			Use: "powershell",
			RunE: func(cmd *cobra.Command, args []string) error {
				return cmd.Root().GenPowerShellCompletion(cmd.OutOrStdout())
			},
		},
	}
}

func newCompletionZshCmd() *cli.Command {
	return &cli.Command{
		UI: cobra.Command{
			Use: "zsh",
			RunE: func(cmd *cobra.Command, args []string) error {
				return cmd.Root().GenZshCompletion(cmd.OutOrStdout())
			},
		},
	}
}
