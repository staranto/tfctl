// Copyright (c) 2025 Steve Taranto staranto@gmail.com.
// SPDX-License-Identifier: Apache-2.0

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

// SvqCommandAction is the action handler for the "svq" subcommand. It lists
// state versions via the active backend, supports --tldr/--schema short-
// circuits, and emits results per common flags.
func SvqCommandAction(ctx context.Context, cmd *cli.Command) error {
	m := GetMeta(cmd)
	log.Debugf("Executing action for %v", m.Args[1:])

	// Bail out early if we're just dumping tldr.
	if ShortCircuitTLDR(ctx, cmd, "svq") {
		return nil
	}

	// Bail out early if we're just dumping the schema.
	if DumpSchemaIfRequested(cmd, reflect.TypeOf(tfe.StateVersion{})) {
		return nil
	}

	attrs := BuildAttrs(cmd, ".id", "serial", "created-at")
	log.Debugf("attrs: %v", attrs)

	// Figure out what type of Backend we're in.
	be, err := backend.NewBackend(ctx, *cmd)
	if err != nil {
		return err
	}
	log.Debugf("typBe: %v", be)

	results, err := be.StateVersions()
	if err != nil {
		return err
	}
	log.Debugf("stateVersions: %v", results)

	if err := EmitJSONAPISlice(results, attrs, cmd); err != nil {
		return err
	}

	return nil
}

// SvqCommandBuilder constructs the cli.Command for "svq", wiring metadata,
// flags, and action/validator handlers.
func SvqCommandBuilder(cmd *cli.Command, meta meta.Meta) *cli.Command {
	return &cli.Command{
		Name:      "svq",
		Usage:     "state version query",
		UsageText: `tfctl svq [RootDir] [options]`,
		Metadata: map[string]any{
			"meta": meta,
		},
		Flags: append([]cli.Flag{
			&cli.BoolFlag{
				Name:    "deep",
				Aliases: []string{"d"},
				Usage:   "enrich results with related data (slower)",
			},
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
			NewHostFlag("svq", meta.Config.Source),
			NewOrgFlag("svq", meta.Config.Source),
			tldrFlag,
			schemaFlag,
			workspaceFlag,
		}, NewGlobalFlags("svq")...),
		Action: func(ctx context.Context, c *cli.Command) error {
			if err := SvqCommandValidator(ctx, c); err != nil {
				return err
			}
			return SvqCommandAction(ctx, c)
		},
	}
}

// SvqCommandValidator performs validation for "svq" and delegates to
// GlobalFlagsValidator.
func SvqCommandValidator(ctx context.Context, cmd *cli.Command) error {
	return GlobalFlagsValidator(ctx, cmd)
}
