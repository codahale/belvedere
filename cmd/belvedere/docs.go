package main

import (
	"github.com/codahale/belvedere/cmd/belvedere/internal/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func newDocsCmd() *cli.Command {
	return &cli.Command{
		UI: cobra.Command{
			Use:    "docs",
			Hidden: true,
		},
		Subcommands: []*cli.Command{
			newDocsManCmd(),
			newDocsMarkdownCmd(),
			newDocsReSTCmd(),
			newDocsYAMLCmd(),
		},
	}
}

func newDocsManCmd() *cli.Command {
	return &cli.Command{
		UI: cobra.Command{
			Use:  "man <dir>",
			Args: cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return doc.GenManTree(cmd.Root(), nil, args[0])
			},
		},
	}
}

func newDocsMarkdownCmd() *cli.Command {
	return &cli.Command{
		UI: cobra.Command{
			Use:  "markdown <dir>",
			Args: cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return doc.GenMarkdownTree(cmd.Root(), args[0])
			},
		},
	}
}

func newDocsReSTCmd() *cli.Command {
	return &cli.Command{
		UI: cobra.Command{
			Use:  "rest <dir>",
			Args: cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return doc.GenReSTTree(cmd.Root(), args[0])
			},
		},
	}
}

func newDocsYAMLCmd() *cli.Command {
	return &cli.Command{
		UI: cobra.Command{
			Use:  "yaml <dir>",
			Args: cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return doc.GenYamlTree(cmd.Root(), args[0])
			},
		},
	}
}
