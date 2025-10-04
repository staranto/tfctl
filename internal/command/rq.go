// Copyright © 2025 Steve Taranto staranto@gmail.com
// SPDX-License-Identifier: MIT

package command

import (
	"context"
	"reflect"

	"github.com/apex/log"
	"github.com/hashicorp/go-tfe"
	altsrc "github.com/urfave/cli-altsrc/v3"
	yaml "github.com/urfave/cli-altsrc/v3/yaml"
	"github.com/urfave/cli/v3"

	"github.com/staranto/tfctlgo/internal/backend"
	"github.com/staranto/tfctlgo/internal/meta"
)

// RqCommandAction is the action handler for the "rq" subcommand. It lists
// runs via the active backend, supports --tldr/--schema short-circuits, and
// emits output per common flags.
func RqCommandAction(ctx context.Context, cmd *cli.Command) error {
	m := GetMeta(cmd)
	log.Debugf("Executing action for %v", m.Args[1:])

	// Bail out early if we're just dumping tldr.
	if ShortCircuitTLDR(ctx, cmd, "rq") {
		return nil
	}

	// Bail out early if we're just dumping the schema.
	if DumpSchemaIfRequested(cmd, reflect.TypeOf(tfe.Run{})) {
		return nil
	}

	attrs := BuildAttrs(cmd, ".id", "created-at", "status")
	log.Debugf("attrs: %v", attrs)

	// Figure out what type of Backend we're in.
	be, err := backend.NewBackend(ctx, *cmd)
	if err != nil {
		return err
	}
	log.Debugf("typBe: %v", be)

	results, err := be.Runs()
	if err != nil {
		return err
	}
	log.Debugf("runs: %v", results)

	if err := EmitJSONAPISlice(results, attrs, cmd); err != nil {
		return err
	}

	return nil
}

// RqCommandBuilder constructs the cli.Command for "rq", wiring metadata,
// flags, and action/validator.
func RqCommandBuilder(cmd *cli.Command, meta meta.Meta) *cli.Command {
	return &cli.Command{
		Name:      "rq",
		Usage:     "run query",
		UsageText: `tfctl rq [RootDir] [options]`,
		Metadata: map[string]any{
			"meta": meta,
		},
		Flags: append([]cli.Flag{
			&cli.IntFlag{
				Name:    "limit",
				Aliases: []string{"l"},
				Usage:   "limit runs returned",
				Sources: cli.NewValueSourceChain(
					yaml.YAML("rq.limit", altsrc.StringSourcer(cfg.Source)),
					yaml.YAML("limit", altsrc.StringSourcer(cfg.Source)),
				),
				Value: 99999,
			},
			NewHostFlag("rq", meta.Config.Source),
			NewOrgFlag("rq", meta.Config.Source),
			tldrFlag,
			schemaFlag,
			workspaceFlag,
		}, NewGlobalFlags("rq")...),
		Action: func(ctx context.Context, c *cli.Command) error {
			if err := RqCommandValidator(ctx, c); err != nil {
				return err
			}
			return RqCommandAction(ctx, c)
		},
	}
}

// RqCommandValidator performs validation for "rq" and delegates to
// GlobalFlagsValidator.
func RqCommandValidator(ctx context.Context, cmd *cli.Command) error {
	return GlobalFlagsValidator(ctx, cmd)
}
