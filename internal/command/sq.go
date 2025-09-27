// Copyright © 2025 Steve Taranto staranto@gmail.com
// SPDX-License-Identifier: MIT

package command

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/apex/log"
	"github.com/staranto/tfctlgo/internal/attrs"
	"github.com/staranto/tfctlgo/internal/backend"
	"github.com/staranto/tfctlgo/internal/config"
	"github.com/staranto/tfctlgo/internal/differ"
	"github.com/staranto/tfctlgo/internal/meta"
	"github.com/staranto/tfctlgo/internal/output"
	"github.com/staranto/tfctlgo/internal/state"
	altsrc "github.com/urfave/cli-altsrc/v3"
	yaml "github.com/urfave/cli-altsrc/v3/yaml"
	"github.com/urfave/cli/v3"
)

func SqCommandAction(ctx context.Context, cmd *cli.Command) error {
	meta := cmd.Metadata["meta"].(meta.Meta)
	log.Debugf("Executing action for %v", meta.Args[1:])

	// Bail out early if we're just dumping tldr.
	if cmd.Bool("tldr") {
		if _, err := exec.LookPath("tldr"); err == nil {
			c := exec.CommandContext(ctx, "tldr", "tfctl", "sq")
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			_ = c.Run()
		}
		return nil
	}

	config.Config.Namespace = "sq"

	// Check to make sure the target directory looks like it might be a legit TF
	// workspace.
	tfConfigFile := fmt.Sprintf("%s/.terraform/terraform.tfstate", meta.RootDir)
	if _, err := os.Stat(tfConfigFile); err != nil {
		return fmt.Errorf("terraform config file not found: %s", tfConfigFile)
	}

	// Figure out what type of Backend we're in.
	be, err := backend.NewBackend(ctx, *cmd)
	if err != nil {
		return err
	}
	log.Debugf("typBe: %v", be)

	// Short circuit --diff mode.
	if cmd.Bool("diff") {
		if _, ok := be.(backend.SelfDiffer); ok {
			states, err := be.(backend.SelfDiffer).DiffStates(ctx, cmd)
			if err != nil {
				log.Errorf("diff error: %v", err)
				return err
			}

			return differ.Diff(ctx, cmd, states)
		} else {
			log.Debug("Backend does not implement SelfDiffer")
		}
	}

	var attrs attrs.AttrList
	//nolint:errcheck
	{
		// mode and type are hidden and used to build the resource attr and for
		// commonly used filtering (eg. --concrete).
		attrs.Set("!.mode")
		attrs.Set("!.type")
		attrs.Set(".resource")
		attrs.Set("id")
		attrs.Set("name")

		extras := cmd.String("attrs")
		if extras != "" {
			attrs.Set(extras)
		}

		attrs.SetGlobalTransformSpec()
	}
	log.Debugf("attrs: %v", attrs)

	var doc []byte
	doc, err = be.State()
	if err != nil {
		return err
	}

	// If the state is encrypted, there's a little more work to do.
	var jsonData map[string]interface{}
	if err := json.Unmarshal(doc, &jsonData); err == nil {
		if _, exists := jsonData["encrypted_data"]; exists {
			// First, look to the flag for passphrase value.
			passphrase := cmd.String("passphrase")

			// Issue 14 - Next look in env and use it if found.
			if passphrase == "" {
				passphrase = os.Getenv("TFCTL_PASSPHRASE")
			}

			// Finally, prompt for passphrase
			if passphrase == "" {
				passphrase, _ = state.GetPassphrase()
			}

			doc, err = state.DecryptOpenTofuState(doc, passphrase)
			if err != nil {
				return fmt.Errorf("failed to decrypt: %w", err)
			}
		}
	}

	var raw bytes.Buffer
	raw.Write(doc)

	output.SliceDiceSpit(raw, attrs, cmd, "", os.Stdout)

	return nil
}

func SqCommandBuilder(cmd *cli.Command, meta meta.Meta) *cli.Command {
	return &cli.Command{
		Name:      "sq",
		Usage:     "state query",
		UsageText: `tfctl sq [RootDir] [options]`,
		Metadata: map[string]any{
			"meta": meta,
		},
		Flags: append([]cli.Flag{
			&cli.BoolFlag{
				Name:    "concrete",
				Aliases: []string{"k"},
				Usage:   "only include concrete resources",
				Value:   false,
			},
			&cli.BoolFlag{
				Name:  "diff",
				Usage: "find difference between state versions",
				Value: false,
			},
			&cli.StringFlag{
				Name:   "diff_filter",
				Hidden: true,
				Value:  "check_results",
				Sources: cli.NewValueSourceChain(
					yaml.YAML("diff_filter", altsrc.StringSourcer(meta.Config.Source)),
				),
			},
			&cli.BoolFlag{
				Name:  "noshort",
				Usage: "include full resource name paths",
				Value: false,
			},
			&cli.StringFlag{
				Name:  "passphrase",
				Usage: "encrypted state passphrase",
			},
			&cli.StringFlag{
				Name:        "sv",
				Usage:       "state version to query",
				Value:       "0",
				HideDefault: true,
			},
			&cli.IntFlag{
				Name:   "limit",
				Hidden: true,
				Usage:  "limit state versions returned",
				Value:  99999,
			},
			NewHostFlag("sq", meta.Config.Source),
			NewOrgFlag("sq", meta.Config.Source),
			tldrFlag,
			workspaceFlag,
		}, NewGlobalFlags("sq")...),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if err := SqCommandValidator(ctx, cmd); err != nil {
				return err
			}
			return SqCommandAction(ctx, cmd)
		},
	}
}

func SqCommandValidator(ctx context.Context, cmd *cli.Command) error {
	return GlobalFlagsValidator(ctx, cmd)
}
