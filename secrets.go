package main

import (
	"context"

	"github.com/codahale/belvedere/pkg/belvedere"
)

type SecretsCmd struct {
	List   SecretsListCmd   `cmd:"" help:"List all secrets."`
	Create SecretsCreateCmd `cmd:"" help:"Create a secret."`
	Grant  SecretsGrantCmd  `cmd:"" help:"Grant access to a secret for an application."`
	Revoke SecretsRevokeCmd `cmd:"" help:"Revoke access to a secret for an application."`
	Update SecretsUpdateCmd `cmd:"" help:"Update a secret."`
	Delete SecretsDeleteCmd `cmd:"" help:"Delete a secret."`
}

type SecretsListCmd struct {
}

func (*SecretsListCmd) Run(ctx context.Context, project belvedere.Project, tables TableWriter) error {
	releases, err := project.Secrets().List(ctx)
	if err != nil {
		return err
	}
	return tables.Print(releases)
}

type SecretsCreateCmd struct {
	Secret   string `arg:"" help:"The secret's name."`
	DataFile string `arg:"" optional:"" help:"File path from which to read secret data. Reads from STDIN if not specified."`
}

func (cmd *SecretsCreateCmd) Run(ctx context.Context, project belvedere.Project, o *Options) error {
	b, err := readFile(ctx, cmd.DataFile)
	if err != nil {
		return err
	}
	return project.Secrets().Create(ctx, cmd.Secret, b, o.DryRun)
}

type SecretsGrantCmd struct {
	Secret string `arg:"" help:"The secret's name."`
	App    string `arg:"" help:"The app's name."`
}

func (cmd *SecretsGrantCmd) Run(ctx context.Context, project belvedere.Project, o *Options) error {
	return project.Secrets().Grant(ctx, cmd.Secret, cmd.App, o.DryRun)
}

type SecretsRevokeCmd struct {
	Secret string `arg:"" help:"The secret's name."`
	App    string `arg:"" help:"The app's name."`
}

func (cmd *SecretsRevokeCmd) Run(ctx context.Context, project belvedere.Project, o *Options) error {
	return project.Secrets().Revoke(ctx, cmd.Secret, cmd.App, o.DryRun)
}

type SecretsUpdateCmd struct {
	Secret   string `arg:"" help:"The secret's name."`
	DataFile string `arg:"" optional:"" help:"File path from which to read secret data. Reads from STDIN if not specified."`
}

func (cmd *SecretsUpdateCmd) Run(ctx context.Context, project belvedere.Project, o *Options) error {
	b, err := readFile(ctx, cmd.DataFile)
	if err != nil {
		return err
	}
	return project.Secrets().Update(ctx, cmd.Secret, b, o.DryRun)
}

type SecretsDeleteCmd struct {
	Secret string `arg:"" help:"The secret's name."`
}

func (cmd *SecretsDeleteCmd) Run(ctx context.Context, project belvedere.Project, o *Options) error {
	return project.Secrets().Delete(ctx, cmd.Secret, o.DryRun)
}
