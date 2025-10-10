// Copyright (c) 2025 Steve Taranto staranto@gmail.com.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/staranto/tfctlgo/internal/cacheutil"
	"github.com/staranto/tfctlgo/internal/command"
	mylog "github.com/staranto/tfctlgo/internal/log"
	"github.com/staranto/tfctlgo/internal/version"
)

var ctx = context.Background()

func main() {
	os.Exit(realMain())
}

func realMain() int {
	mylog.InitLogger()

	args := os.Args

	// Best-effort: pre-create cache directory when caching is enabled.
	if _, ok, err := cacheutil.EnsureBaseDir(); err != nil && ok {
		// Non-fatal: print to stderr and continue.
		fmt.Fprintln(os.Stderr, err)
	}

	// TODO Let urfave do this
	// Short-circuit --version/-v.
	for _, a := range args {
		if a == "--version" || a == "-v" {
			fmt.Println(version.Version)
			return 0
		}
	}

	app, err := command.InitApp(ctx, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if err := app.Run(ctx, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	return 0
}
