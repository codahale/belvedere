package main

import (
	"context"

	"github.com/codahale/belvedere/pkg/belvedere/secrets"
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

func (*SecretsListCmd) Run(ctx context.Context, o *Options) error {
	s, err := secrets.NewService(ctx, o.Project, o.DryRun)
	if err != nil {
		return err
	}

	releases, err := s.List(ctx)
	if err != nil {
		return err
	}
	return o.printTable(releases)
}

type SecretsCreateCmd struct {
	Secret   string `arg:"" help:"The secret's name."`
	DataFile string `arg:"" optional:"" help:"File path from which to read secret data. Reads from STDIN if not specified."`
}

func (cmd *SecretsCreateCmd) Run(ctx context.Context, o *Options) error {
	s, err := secrets.NewService(ctx, o.Project, o.DryRun)
	if err != nil {
		return err
	}

	b, err := readFile(ctx, cmd.DataFile)
	if err != nil {
		return err
	}
	return s.Create(ctx, cmd.Secret, b)
}

type SecretsGrantCmd struct {
	Secret string `arg:"" help:"The secret's name."`
	App    string `arg:"" help:"The app's name."`
}

func (cmd *SecretsGrantCmd) Run(ctx context.Context, o *Options) error {
	s, err := secrets.NewService(ctx, o.Project, o.DryRun)
	if err != nil {
		return err
	}

	return s.Grant(ctx, cmd.Secret, cmd.App)
}

type SecretsRevokeCmd struct {
	Secret string `arg:"" help:"The secret's name."`
	App    string `arg:"" help:"The app's name."`
}

func (cmd *SecretsRevokeCmd) Run(ctx context.Context, o *Options) error {
	s, err := secrets.NewService(ctx, o.Project, o.DryRun)
	if err != nil {
		return err
	}

	return s.Revoke(ctx, cmd.Secret, cmd.App)
}

type SecretsUpdateCmd struct {
	Secret   string `arg:"" help:"The secret's name."`
	DataFile string `arg:"" optional:"" help:"File path from which to read secret data. Reads from STDIN if not specified."`
}

func (cmd *SecretsUpdateCmd) Run(ctx context.Context, o *Options) error {
	s, err := secrets.NewService(ctx, o.Project, o.DryRun)
	if err != nil {
		return err
	}

	b, err := readFile(ctx, cmd.DataFile)
	if err != nil {
		return err
	}
	return s.Update(ctx, cmd.Secret, b)
}

type SecretsDeleteCmd struct {
	Secret string `arg:"" help:"The secret's name."`
}

func (cmd *SecretsDeleteCmd) Run(ctx context.Context, o *Options) error {
	s, err := secrets.NewService(ctx, o.Project, o.DryRun)
	if err != nil {
		return err
	}

	return s.Delete(ctx, cmd.Secret)
}
