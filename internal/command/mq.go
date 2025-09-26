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
	"github.com/staranto/tfctlgo/internal/backend/remote"
	"github.com/staranto/tfctlgo/internal/meta"
	"github.com/staranto/tfctlgo/internal/output"
	"github.com/urfave/cli/v3"
)

func MqCommandAction(ctx context.Context, cmd *cli.Command) error {
	meta := cmd.Metadata["meta"].(meta.Meta)
	log.Debugf("Executing action for %v", meta.Args[1:])

	// Bail out early if we're just dumping tldr.
	if cmd.Bool("tldr") {
		if _, err := exec.LookPath("tldr"); err == nil {
			cmd := exec.CommandContext(ctx, "tldr", "tfctl", "mq")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			_ = cmd.Run()
		}
		return nil
	}

	// Bail out early if we're just dumping the schema.
	if cmd.Bool("schema") {
		output.DumpSchema("", reflect.TypeOf(tfe.RegistryModule{}))
		return nil
	}

	var attrs attrs.AttrList
	//nolint:errcheck
	{
		attrs.Set(".id")
		attrs.Set("name")

		extras := cmd.String("attrs")
		if extras != "" {
			attrs.Set(extras)
		}

		attrs.SetGlobalTransformSpec()
	}
	log.Debugf("attrs: %v", attrs)

	be, err := remote.NewBackendRemote(ctx, cmd, remote.BuckNaked())
	if err != nil {
		return err
	}
	log.Debugf("be: %v", be)

	client, err := be.Client()
	if err != nil {
		return err
	}
	log.Debugf("client: %v", client.BaseURL())

	options := tfe.RegistryModuleListOptions{
		ListOptions: tfe.ListOptions{PageNumber: 1, PageSize: 100},
	}

	var results []*tfe.RegistryModule

	// Paginate through the dataset
	for {
		page, err := client.RegistryModules.List(ctx, cmd.String("org"), &options)
		if err != nil {
			return fmt.Errorf("failed to list registry modules: %w", err)
		}

		results = append(results, page.Items...)
		log.Debugf("page: %d, total: %d", page.CurrentPage, len(results))

		if page.Pagination.NextPage == 0 {
			break
		}
		options.ListOptions.PageNumber++
	}

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

func MqCommandBuilder(cmd *cli.Command, meta meta.Meta, globalFlags []cli.Flag) *cli.Command {
	return &cli.Command{
		Name:      "mq",
		Usage:     "module registry query",
		UsageText: `tfctl mq [RootDir] [options]`,
		Metadata: map[string]any{
			"meta": meta,
		},
		Flags: append([]cli.Flag{
			NewHostFlag("mq", meta.Config.Source),
			NewOrgFlag("mq", meta.Config.Source),
			tldrFlag,
			schemaFlag,
		}, NewGlobalFlags("mq")...),
		Action: func(ctx context.Context, c *cli.Command) error {
			if err := MqCommandValidator(ctx, c, globalFlags); err != nil {
				return err
			}
			return MqCommandAction(ctx, c)
		},
	}
}

func MqCommandValidator(ctx context.Context, cmd *cli.Command, globalFlags []cli.Flag) error {
	return GlobalFlagsValidator(ctx, cmd, globalFlags)
}
