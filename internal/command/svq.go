// Copyright © 2025 Steve Taranto staranto@gmail.com
// SPDX-License-Identifier: MIT

package command

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"reflect"

	"github.com/apex/log"
	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/jsonapi"
	"github.com/staranto/tfctlgo/internal/attrs"
	"github.com/staranto/tfctlgo/internal/backend"
	"github.com/staranto/tfctlgo/internal/meta"
	"github.com/staranto/tfctlgo/internal/output"
	altsrc "github.com/urfave/cli-altsrc/v3"
	yaml "github.com/urfave/cli-altsrc/v3/yaml"
	"github.com/urfave/cli/v3"
)

func SvqCommandAction(ctx context.Context, cmd *cli.Command) error {
	meta := cmd.Metadata["meta"].(meta.Meta)
	log.Debugf("Executing action for %v", meta.Args[1:])

	// Bail out early if we're just dumping tldr.
	if cmd.Bool("tldr") {
		if _, err := exec.LookPath("tldr"); err == nil {
			c := exec.CommandContext(ctx, "tldr", "tfctl", "svq")
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			_ = c.Run()
		}
		return nil
	}

	// TODO Add examples.

	// Bail out early if we're just dumping the schema.
	if cmd.Bool("schema") {
		output.DumpSchema("", reflect.TypeOf(tfe.StateVersion{}))
		return nil
	}

	var attrs attrs.AttrList
	//nolint:errcheck
	{
		attrs.Set(".id")
		attrs.Set("serial")
		attrs.Set("created-at")

		extras := cmd.String("attrs")
		if extras != "" {
			attrs.Set(extras)
		}

		attrs.SetGlobalTransformSpec()
	}
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

	// Marshal into a JSON document so we can slice and dice some more.  Note that
	// we're using jsonapi, which will use the StructField tags as the keys of the
	// JSON document.
	var raw bytes.Buffer
	if err := jsonapi.MarshalPayload(&raw, results); err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	output.SliceDiceSpit(raw, attrs, cmd, "data", os.Stdout)

	return nil
}

func SvqCommandBuilder(cmd *cli.Command, meta meta.Meta, globalFlags []cli.Flag) *cli.Command {
	return &cli.Command{
		Name:      "svq",
		Usage:     "state version query",
		UsageText: `tfctl svq [RootDir] [options]`,
		Metadata: map[string]any{
			"meta": meta,
		},
		Flags: append([]cli.Flag{
			&cli.IntFlag{
				Name:    "limit",
				Aliases: []string{"l"},
				Usage:   "limit state versions returned",
				Value:   99999,
				Sources: cli.NewValueSourceChain(
					yaml.YAML("limit", altsrc.StringSourcer(meta.Config.Source)),
				),
			},
			NewHostFlag("svq", meta.Config.Source),
			NewOrgFlag("svq", meta.Config.Source),
			tldrFlag,
			schemaFlag,
			workspaceFlag,
		}, globalFlags...),
		Action: func(ctx context.Context, c *cli.Command) error {
			if err := SvqCommandValidator(ctx, c, globalFlags); err != nil {
				return err
			}
			return SvqCommandAction(ctx, c)
		},
	}
}

func SvqCommandValidator(ctx context.Context, cmd *cli.Command, globalFlags []cli.Flag) error {
	return GlobalFlagsValidator(ctx, cmd, globalFlags)
}
