// Copyright (c) 2025 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/apex/log"

	"github.com/staranto/tfctlgo/internal/cacheutil"
	"github.com/staranto/tfctlgo/internal/command"
	"github.com/staranto/tfctlgo/internal/config"
	mylog "github.com/staranto/tfctlgo/internal/log"
	"github.com/staranto/tfctlgo/internal/util"
	"github.com/staranto/tfctlgo/internal/version"
)

var ctx = context.Background()

func main() {
	os.Exit(realMain())
}

func realMain() int {
	mylog.InitLogger()

	args := os.Args

	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "No command specified.")
		args = append(args, "--help")
	} else {
		args = mangleArguments(args)
	}

	// Short-circuit --version/-v.
	for _, a := range args {
		if a == "--version" || a == "-v" {
			fmt.Println(version.Version)
			return 0
		}
	}

	// Best-effort: pre-create cache directory when caching is enabled.
	if _, ok, err := cacheutil.EnsureBaseDir(); err != nil && ok {
		// Non-fatal: print to stderr and continue.
		fmt.Fprintln(os.Stderr, err)
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

func mangleArguments(args []string) []string {
	// We know the first two args are going to be the executable and command.
	preamble := make([]string, 2)
	copy(preamble, args[:2])

	// Short-circuit for --help/-h. If help is requested, just keep the preamble
	// and add --help flag.
	for _, a := range args {
		if a == "--help" || a == "-h" {
			return append(preamble, "--help")
		}
	}

	// And the next arg might be a root dir
	rootDir, _ := os.Getwd()
	argStartIdx := 2
	if len(args) > 2 {
		if _, _, err := util.ParseRootDir(args[2]); err == nil {
			rootDir = args[2]
			argStartIdx = 3
		}
	}

	defaultSet := "@defaults"

	// Scan through the args. If there is no @set, just use it and ignore this
	// default.
	for _, a := range args {
		if strings.HasPrefix(a, "@") {
			defaultSet = ""
			break
		}
	}

	// Now combine them back together.
	workingArgs := append(preamble, rootDir) //nolint
	if defaultSet != "" {
		workingArgs = append(workingArgs, defaultSet)
	}

	if argStartIdx < len(args) {
		workingArgs = append(workingArgs, args[argStartIdx:]...)
	}

	args = workingArgs

	// Now scan through args and if there is not a @set, insert @defaults after
	idx := 2
	set := "defaults"
	// See if there is a @set specified. If so, that becomes are insertion point
	// and the @set entry is removed from args.
	for i, a := range args[idx:] {
		if strings.HasPrefix(a, "@") {
			set = a[1:]
			idx += i
			args = append(args[:idx], args[idx+1:]...)
			break
		}
	}

	setArgs, _ := config.GetStringSlice(args[1] + "." + set)
	for _, arg := range setArgs {
		parts := strings.Fields(arg)
		args = append(args[:idx], append(parts, args[idx:]...)...)
		idx += len(parts)
	}

	log.Debugf("idx=%d, set=%s, args=%v", idx, set, args)
	return args
}
