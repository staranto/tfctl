// Copyright © 2025 Steve Taranto staranto@gmail.com
// SPDX-License-Identifier: MIT

package command

import (
	"context"
	"fmt"
	"reflect"

	"github.com/apex/log"
	"github.com/hashicorp/go-tfe"
	"github.com/staranto/tfctlgo/internal/backend/remote"
	"github.com/staranto/tfctlgo/internal/meta"
	"github.com/urfave/cli/v3"
)

// WqCommandAction is the action handler for the "wq" subcommand. It lists
// workspaces for the selected organization, supports --tldr/--schema short-
// circuits, and emits results per common flags.
func WqCommandAction(ctx context.Context, cmd *cli.Command) error {
	m := GetMeta(cmd)
	log.Debugf("Executing action for %v", m.Args[1:])

	// Bail out early if we're just dumping tldr.
	if ShortCircuitTLDR(ctx, cmd, "wq") {
		return nil
	}

	// Bail out early if we're just dumping the schema.
	if DumpSchemaIfRequested(cmd, reflect.TypeOf(tfe.Workspace{})) {
		return nil
	}

	attrs := BuildAttrs(cmd, ".id", "name")
	log.Debugf("attrs: %v", attrs)

	//be, _ := remote.NewConfigRemote(remote.BuckNaked())
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

	options := tfe.WorkspaceListOptions{
		ListOptions: tfe.ListOptions{PageNumber: 1, PageSize: 100},
	}

	var results []*tfe.Workspace

	// Paginate through the dataset
	for {
		page, err := client.Workspaces.List(ctx, cmd.String("org"), &options)
		if err != nil {
			return fmt.Errorf("failed to list workspaces: %w", err)
		}

		results = append(results, page.Items...)
		log.Debugf("page: %d, total: %d", page.CurrentPage, len(results))

		if page.Pagination.NextPage == 0 {
			break
		}
		options.ListOptions.PageNumber++
	}

	if err := EmitJSONAPISlice(results, attrs, cmd); err != nil {
		return err
	}

	return nil
}

// WqCommandBuilder constructs the cli.Command for "wq", wiring metadata,
// flags, and action/validator handlers.
func WqCommandBuilder(cmd *cli.Command, meta meta.Meta) *cli.Command {
	return &cli.Command{
		Name:      "wq",
		Usage:     "workspace query",
		UsageText: `tfctl wq [RootDir] [options]`,
		Metadata: map[string]any{
			"meta": meta,
		},
		Flags: append([]cli.Flag{
			NewHostFlag("wq", meta.Config.Source),
			NewOrgFlag("wq", meta.Config.Source),
			tldrFlag,
			schemaFlag,
		}, NewGlobalFlags("wq")...),
		Action: func(ctx context.Context, c *cli.Command) error {
			if err := WqCommandValidator(ctx, c); err != nil {
				return err
			}
			return WqCommandAction(ctx, c)
		},
	}
}

// WqCommandValidator performs validation for "wq" and delegates to
// GlobalFlagsValidator.
func WqCommandValidator(ctx context.Context, cmd *cli.Command) error {
	return GlobalFlagsValidator(ctx, cmd)
}
