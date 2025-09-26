// Copyright © 2025 Steve Taranto staranto@gmail.com
// SPDX-License-Identifier: MIT
package command

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/staranto/tfctlgo/internal/config"
	"github.com/staranto/tfctlgo/internal/meta"
	"github.com/staranto/tfctlgo/internal/util"
	"github.com/urfave/cli/v3"
)

func InitApp(ctx context.Context, args []string) (*cli.Command, error) {

	// Save the CWD at startup and then defer restoring it so we're tidy.
	sd, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(sd); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to restore directory: %v\n", err)
		}
	}()

	// The arg[1] immediately following the binary (arg[0]) is the tfctl
	// subcommand and also represents the namespace key to be used when retrieving
	// config values. arg[1] could be -h/--help, so ignore it if it appears to be
	// a flag.
	var ns string
	if len(args) > 1 && !strings.HasPrefix(args[1], "-") {
		ns = args[1]
	}

	cfg, _ := config.Load(ns)
	meta := meta.Meta{
		Args:        args,
		Config:      cfg,
		Context:     ctx,
		StartingDir: sd,
	}

	// See if the arg immediately following the command might be a directory.
	// This is determined by whether or not it begins with - or --.  If it does,
	// it's a flag and the CWD directory is the starting directory.  If it's not,
	// we assume we have a directory spec of some sort and need to parse it more.
	if len(args) > 2 && !strings.HasPrefix(args[2], "-") {
		if wd, env, err := util.ParseRootDir(args[2]); err == nil {
			meta.RootDir = wd
			meta.Env = env
		} else {
			return nil, fmt.Errorf("failed to parse rootDir (%s): %w", args[2], err)
		}
	} else {
		meta.RootDir = sd
	}

	app := &cli.Command{
		Name:  "tfctl",
		Usage: "Terraform Control",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "version",
				Aliases:     []string{"v"},
				Usage:       "tfctl version info",
				HideDefault: true,
			},
		},
	}

	app.Commands = append(app.Commands,
		MqCommandBuilder(app, meta, GlobalFlags),
		OqCommandBuilder(app, meta, GlobalFlags),
		PqCommandBuilder(app, meta, GlobalFlags),
		RqCommandBuilder(app, meta, GlobalFlags),
		SiCommandBuilder(app, meta, GlobalFlags),
		SqCommandBuilder(app, meta, GlobalFlags),
		SvqCommandBuilder(app, meta, GlobalFlags),
		WqCommandBuilder(app, meta, GlobalFlags),
	)

	// Make sure flags are sorted for the --help text.
	for _, cmd := range app.Commands {
		sort.Slice(cmd.Flags, func(i, j int) bool {
			return cmd.Flags[i].Names()[0] < cmd.Flags[j].Names()[0]
		})
	}

	return app, nil
}
