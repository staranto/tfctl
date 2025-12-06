// Copyright (c) 2025 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/staranto/tfctlgo/internal/cacheutil"
	"github.com/staranto/tfctlgo/internal/command"
	"github.com/staranto/tfctlgo/internal/config"
	"github.com/staranto/tfctlgo/internal/log"
	"github.com/staranto/tfctlgo/internal/util"
	"github.com/staranto/tfctlgo/internal/version"
)

var ctx = context.Background()

func main() {
	os.Exit(realMain())
}

func realMain() int {
	log.InitLogger()

	args := os.Args
	log.Debugf("args captured: args=%v", args)

	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "No command specified.")
		args = append(args, "--help")
		log.Debugf("help added: args=%v", args)
	} else {
		args = mangleArguments(args)
		log.Debugf("args mangled: args=%v", args)
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
		log.Debugf("cache ensure err: err=%v", err)
	}

	app, err := command.InitApp(ctx, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		log.Debugf("app init err: err=%v", err)
		return 1
	}

	if err := app.Run(ctx, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		log.Debugf("app run err: err=%v", err)
		return 2
	}

	return 0
}

func mangleArguments(args []string) []string {
	log.Debugf("mangle start: args=%v", args)
	// We know the first two args are going to be the executable and command.
	preamble := make([]string, 2)
	copy(preamble, args[:2])
	log.Debugf("preamble set: preamble=%v", preamble)

	// Short-circuit for --help/-h. If help is requested, just keep the preamble
	// and add --help flag.
	for _, a := range args {
		if a == "--help" || a == "-h" {
			log.Debugf("help short: args=%v", append(preamble, "--help"))
			return append(preamble, "--help")
		}
	}

	// And the next arg might be a root dir
	rootDir, _ := os.Getwd()
	argStartIdx := 2
	// Special-case the 'completion' and 'ps' commands which take a plain
	// positional argument (e.g., 'bash' or 'zsh' for completion, plan file
	// for ps).
	isSpecial := args[1] == "ps" || args[1] == "completion"

	if !isSpecial && len(args) > 2 {
		if _, _, err := util.ParseRootDir(args[2]); err == nil {
			rootDir = args[2]
			argStartIdx = 3
		}
	}
	log.Debugf("root dir set: rootDir=%s, argStartIdx=%d", rootDir, argStartIdx)

	// This first scan through args is going to see if we should keep @defaults or
	// not.  It should only be kept when there is not a explicitly provided @set.
	defaultSet := "@defaults"
	for _, a := range args {
		if strings.HasPrefix(a, "@") {
			defaultSet = ""
			break
		}
	}
	log.Debugf("default set: defaultSet=%s", defaultSet)

	// Now combine them back together.
	var workingArgs []string
	if isSpecial {
		workingArgs = make([]string, 2)
		copy(workingArgs, preamble)
	} else {
		workingArgs = append(preamble, rootDir) //nolint
	}

	if defaultSet != "" {
		workingArgs = append(workingArgs, defaultSet)
	}

	if argStartIdx < len(args) {
		workingArgs = append(workingArgs, args[argStartIdx:]...)
	}
	args = workingArgs
	log.Debugf("working args: args=%v", args)

	// Look for an explicit @set argument.
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
	log.Debugf("set found: set=%s, idx=%d", set, idx)

	setArgs, _ := config.GetStringSlice(args[1] + "." + set)
	log.Debugf("set args got: setArgs=%v", setArgs)
	for _, arg := range setArgs {
		parts := strings.Fields(arg)
		args = append(args[:idx], append(parts, args[idx:]...)...)
		idx += len(parts)
	}
	log.Debugf("set expanded: idx=%d, set=%s, args=%v", idx, set, args)

	return args
}
