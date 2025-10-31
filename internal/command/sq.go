// Copyright (c) 2025 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/apex/log"
	altsrc "github.com/urfave/cli-altsrc/v3"
	yaml "github.com/urfave/cli-altsrc/v3/yaml"
	"github.com/urfave/cli/v3"

	"github.com/staranto/tfctlgo/internal/backend"
	"github.com/staranto/tfctlgo/internal/config"
	"github.com/staranto/tfctlgo/internal/differ"
	"github.com/staranto/tfctlgo/internal/meta"
	"github.com/staranto/tfctlgo/internal/output"
	"github.com/staranto/tfctlgo/internal/state"
)

// sqCommandAction is the action handler for the "sq" subcommand. It reads
// Terraform state (including optional decryption), supports --tldr short-
// circuit, and emits results per common flags.
func sqCommandAction(ctx context.Context, cmd *cli.Command) error {
	m := GetMeta(cmd)
	log.Debugf("Executing action for %v", m.Args[1:])

	// Bail out early if we're just dumping tldr.
	if ShortCircuitTLDR(ctx, cmd, "sq") {
		return nil
	}

	config.Config.Namespace = "sq"

	// Figure out what type of Backend we're in.
	be, err := backend.NewBackend(ctx, *cmd)
	if err != nil {
		return err
	}
	log.Debugf("typBe: %v", be)

	// Short circuit --diff mode.
	if cmd.Bool("diff") {
		if _, ok := be.(backend.SelfDiffer); ok {
			states, diffErr := be.(backend.SelfDiffer).DiffStates(ctx, cmd)
			if diffErr != nil {
				log.Errorf("diff error: %v", diffErr)
				return diffErr
			}

			return differ.Diff(ctx, cmd, states)
		} else {
			log.Debug("Backend does not implement SelfDiffer")
		}
	}

	attrs := BuildAttrs(cmd, "!.mode", "!.type", ".resource", "id", "name")
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

	postProcess := func(dataset []map[string]interface{}) error {
		if cmd.Bool("chop") {
			chopPrefix(dataset)
		}

		return nil
	}

	output.SliceDiceSpit(raw, attrs, cmd, "", os.Stdout, postProcess)

	return nil
}

// sqCommandBuilder constructs the cli.Command for "sq", wiring metadata,
// flags, and action/validator handlers.
func sqCommandBuilder(meta meta.Meta) *cli.Command {
	return &cli.Command{
		Name:      "sq",
		Usage:     "state query",
		UsageText: `tfctl sq [RootDir] [options]`,
		Metadata: map[string]any{
			"meta": meta,
		},
		Flags: append([]cli.Flag{
			&cli.BoolFlag{
				Name:  "chop",
				Usage: "chop common resource prefix from names",
				Sources: cli.NewValueSourceChain(
					yaml.YAML("sq.chop", altsrc.StringSourcer(meta.Config.Source)),
				),
				Value: false,
			},
			&cli.BoolFlag{
				Name:    "concrete",
				Aliases: []string{"k"},
				Usage:   "only include concrete resources",
				Sources: cli.NewValueSourceChain(
					yaml.YAML("sq.concrete", altsrc.StringSourcer(cfg.Source)),
				),
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "diff",
				Usage: "find difference between state versions",
				Value: false,
			},
			&cli.StringFlag{
				Name:   "diff_filter",
				Hidden: true,
				Sources: cli.NewValueSourceChain(
					yaml.YAML("sq.diff_filter", altsrc.StringSourcer(meta.Config.Source)),
				),
				Value: "check_results",
			},
			&cli.IntFlag{
				Name:   "limit",
				Hidden: true,
				Usage:  "limit state versions returned",
				Sources: cli.NewValueSourceChain(
					yaml.YAML("sq.limit", altsrc.StringSourcer(cfg.Source)),
					yaml.YAML("limit", altsrc.StringSourcer(cfg.Source)),
				),
				Value: 99999,
			},
			&cli.BoolWithInverseFlag{
				Name:  "short",
				Usage: "include full resource name paths",
				Sources: cli.NewValueSourceChain(
					yaml.YAML("sq.short", altsrc.StringSourcer(meta.Config.Source)),
				),
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
			// We don't want sq to get default host and org values from the config.
			// Instead, we'll depend on the backend or, in exceptional cases, explicit
			// --host and --org flags.
			NewHostFlag("sq"),
			NewOrgFlag("sq"),
			tldrFlag,
			workspaceFlag,
		}, NewGlobalFlags("sq")...),
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			// If --chop is set, --short must not be set.
			if cmd.Bool("chop") {
				_ = cmd.Set("short", "false")
			}

			return ctx, GlobalFlagsValidator(ctx, cmd)
		},
		Action: sqCommandAction,
	}
}

// chopPrefix finds common leading dot-delimited segments in the
// resource attribute of dataset values. If at least 50% of entries share
// at least 2 common leading segments, those segments (and the trailing dot)
// are removed and replaced with ".. ".
func chopPrefix(dataset []map[string]interface{}) {
	// Bail out early if there is no data.
	if len(dataset) == 0 {
		return
	}

	// Collect all resource values with their indices.
	type resourceEntry struct {
		idx   int
		value string
	}

	var resourceValues []resourceEntry
	for i, entry := range dataset {
		if val, ok := entry["resource"]; ok {
			if str, ok := val.(string); ok {
				resourceValues = append(resourceValues, resourceEntry{idx: i, value: str})
			}
		}
	}

	// Calculate the 50% threshold.
	threshold := (len(resourceValues) + 1) / 2

	// Split each value by dots and find common leading segments.
	type segmentedValue struct {
		idx      int
		value    string
		segments []string
	}

	var segmented []segmentedValue
	maxSegments := 0
	for _, rv := range resourceValues {
		segs := strings.Split(rv.value, ".")
		segmented = append(segmented, segmentedValue{idx: rv.idx, value: rv.value, segments: segs})
		if len(segs) > maxSegments {
			maxSegments = len(segs)
		}
	}

	// Find the longest common prefix of segments that appears in at least 50%.
	var commonSegments []string
	for segIdx := 0; segIdx < maxSegments; segIdx++ {
		// Count how many values have a segment at this position and what it is.
		segmentCounts := make(map[string]int)
		for _, sv := range segmented {
			if segIdx < len(sv.segments) {
				segmentCounts[sv.segments[segIdx]]++
			}
		}

		// Find the most common segment at this position.
		bestSegment, bestCount := findBestSegment(segmentCounts)

		// Bail out if the threshold is not met.
		if bestCount < threshold {
			break
		}

		// Otherwise, add it and keep going.
		commonSegments = append(commonSegments, bestSegment)
	}

	// If we don't have at least 2 commonSegments, bail out. There's nothing
	// worth chopping.
	if len(commonSegments) < 2 {
		return
	}

	// Now, chop the common prefix from all matching resource values.
	prefixToRemove := strings.Join(commonSegments, ".") + "."
	for _, sv := range segmented {
		if strings.HasPrefix(sv.value, prefixToRemove) {
			dataset[sv.idx]["resource"] = ".." + sv.value[len(prefixToRemove):]
		}
	}
}

func findBestSegment(segmentCounts map[string]int) (string, int) {
	var bestSegment string
	var bestCount int
	for seg, count := range segmentCounts {
		if count > bestCount {
			bestSegment = seg
			bestCount = count
		}
	}
	return bestSegment, bestCount
}
