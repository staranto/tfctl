// Copyright (c) 2025 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"context"
	"reflect"

	"github.com/hashicorp/go-tfe"
	altsrc "github.com/urfave/cli-altsrc/v3"
	yaml "github.com/urfave/cli-altsrc/v3/yaml"
	"github.com/urfave/cli/v3"

	"github.com/staranto/tfctlgo/internal/meta"
)

// RqCommandAction is the action handler for the "rq" subcommand. It lists
// runs via the active backend, supports --tldr/--schema short-circuits, and
// emits output per common flags.
func RqCommandAction(ctx context.Context, cmd *cli.Command) error {
	be, err := InitLocalBackendQuery(ctx, cmd)
	if err != nil {
		return err
	}

	runner := &QueryActionRunner[*tfe.Run]{
		CommandName:  "rq",
		SchemaType:   reflect.TypeOf(tfe.Run{}),
		DefaultAttrs: []string{".id", "created-at", "status"},
		FetchFn: func(ctx context.Context, cmd *cli.Command) (
			[]*tfe.Run,
			error,
		) {
			return be.Runs()
		},
	}
	return runner.Run(ctx, cmd)
}

// RqCommandBuilder constructs the cli.Command for "rq", wiring metadata,
// flags, and action/validator.
func RqCommandBuilder(cmd *cli.Command, meta meta.Meta) *cli.Command {
	return (&QueryCommandBuilder{
		Name:      "rq",
		Usage:     "run query",
		UsageText: `tfctl rq [RootDir] [options]`,
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "limit",
				Aliases: []string{"l"},
				Usage:   "limit runs returned",
				Sources: cli.NewValueSourceChain(
					yaml.YAML("rq.limit", altsrc.StringSourcer(meta.Config.Source)),
					yaml.YAML("limit", altsrc.StringSourcer(meta.Config.Source)),
				),
				Value: 99999,
			},
			NewHostFlag("rq", meta.Config.Source),
			NewOrgFlag("rq", meta.Config.Source),
			workspaceFlag,
		},
		Action: RqCommandAction,
		Meta:   meta,
	}).Build()
}

// RqCommandValidator performs validation for "rq" and delegates to
// GlobalFlagsValidator.
func RqCommandValidator(ctx context.Context, cmd *cli.Command) error {
	return GlobalFlagsValidator(ctx, cmd)
}
