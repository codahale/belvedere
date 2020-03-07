package secrets

import (
	"context"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cmd"
	"github.com/codahale/belvedere/pkg/belvedere"
)

type RootCmd struct {
	List   ListCmd   `cmd:"" help:"List all secrets."`
	Create CreateCmd `cmd:"" help:"Create a secret."`
	Grant  GrantCmd  `cmd:"" help:"Grant access to a secret for an application."`
	Revoke RevokeCmd `cmd:"" help:"Revoke access to a secret for an application."`
	Update UpdateCmd `cmd:"" help:"Update a secret."`
	Delete DeleteCmd `cmd:"" help:"Delete a secret."`
}

type ListCmd struct {
}

func (*ListCmd) Run(ctx context.Context, project belvedere.Project, tables cmd.TableWriter) error {
	releases, err := project.Secrets().List(ctx)
	if err != nil {
		return err
	}
	return tables.Print(releases)
}

type CreateCmd struct {
	Secret            string `arg:"" help:"The secret's name."`
	DataFile          string `arg:"" default:"-" optional:"" help:"File path from which to read secret data. Reads from STDIN if not specified."`
	cmd.ModifyOptions `embed:""`
}

func (cmd *CreateCmd) Run(ctx context.Context, project belvedere.Project, fr cmd.FileReader) error {
	b, err := fr.Read(cmd.DataFile)
	if err != nil {
		return err
	}
	return project.Secrets().Create(ctx, cmd.Secret, b, cmd.DryRun)
}

type GrantCmd struct {
	Secret            string `arg:"" help:"The secret's name."`
	App               string `arg:"" help:"The app's name."`
	cmd.ModifyOptions `embed:""`
}

func (cmd *GrantCmd) Run(ctx context.Context, project belvedere.Project) error {
	return project.Secrets().Grant(ctx, cmd.Secret, cmd.App, cmd.DryRun)
}

type RevokeCmd struct {
	Secret            string `arg:"" help:"The secret's name."`
	App               string `arg:"" help:"The app's name."`
	cmd.ModifyOptions `embed:""`
}

func (cmd *RevokeCmd) Run(ctx context.Context, project belvedere.Project) error {
	return project.Secrets().Revoke(ctx, cmd.Secret, cmd.App, cmd.DryRun)
}

type UpdateCmd struct {
	Secret            string `arg:"" help:"The secret's name."`
	DataFile          string `arg:"" default:"-" optional:"" help:"File path from which to read secret data. Reads from STDIN if not specified."`
	cmd.ModifyOptions `embed:""`
}

func (cmd *UpdateCmd) Run(ctx context.Context, project belvedere.Project, fr cmd.FileReader) error {
	b, err := fr.Read(cmd.DataFile)
	if err != nil {
		return err
	}
	return project.Secrets().Update(ctx, cmd.Secret, b, cmd.DryRun)
}

type DeleteCmd struct {
	Secret            string `arg:"" help:"The secret's name."`
	cmd.ModifyOptions `embed:""`
}

func (cmd *DeleteCmd) Run(ctx context.Context, project belvedere.Project) error {
	return project.Secrets().Delete(ctx, cmd.Secret, cmd.DryRun)
}
