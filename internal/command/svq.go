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

// SvqCommandAction is the action handler for the "svq" subcommand. It lists
// state versions via the active backend, supports --tldr/--schema short-
// circuits, and emits results per common flags.
func SvqCommandAction(ctx context.Context, cmd *cli.Command) error {
	be, err := InitLocalBackendQuery(ctx, cmd)
	if err != nil {
		return err
	}

	runner := &QueryActionRunner[*tfe.StateVersion]{
		CommandName:  "svq",
		SchemaType:   reflect.TypeOf(tfe.StateVersion{}),
		DefaultAttrs: []string{".id", "serial", "created-at"},
		FetchFn: func(ctx context.Context, cmd *cli.Command) (
			[]*tfe.StateVersion,
			error,
		) {
			return be.StateVersions()
		},
	}
	return runner.Run(ctx, cmd)
}

// SvqCommandBuilder constructs the cli.Command for "svq", wiring metadata,
// flags, and action/validator handlers.
func SvqCommandBuilder(cmd *cli.Command, meta meta.Meta) *cli.Command {
	return (&QueryCommandBuilder{
		Name:      "svq",
		Usage:     "state version query",
		UsageText: `tfctl svq [RootDir] [options]`,
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "limit",
				Aliases: []string{"l"},
				Usage:   "limit state versions returned",
				Sources: cli.NewValueSourceChain(
					yaml.YAML("svq.limit", altsrc.StringSourcer(meta.Config.Source)),
					yaml.YAML("limit", altsrc.StringSourcer(meta.Config.Source)),
				),
				Value: 99999,
			},
			NewHostFlag("svq"),
			NewOrgFlag("svq"),
			workspaceFlag,
		},
		Action: SvqCommandAction,
		Meta:   meta,
	}).Build()
}

// SvqCommandValidator performs validation for "svq" and delegates to
// GlobalFlagsValidator.
func SvqCommandValidator(ctx context.Context, cmd *cli.Command) error {
	return GlobalFlagsValidator(ctx, cmd)
}
