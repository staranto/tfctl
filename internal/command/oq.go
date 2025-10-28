// Copyright (c) 2025 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"context"
	"reflect"

	"github.com/hashicorp/go-tfe"
	"github.com/urfave/cli/v3"

	"github.com/staranto/tfctlgo/internal/backend/remote"
	"github.com/staranto/tfctlgo/internal/meta"
)

// oqCommandAction is the action handler for the "oq" subcommand. It lists
// organizations from the configured host, supports --tldr/--schema
// short-circuit behavior, and emits output per common flags.
func oqCommandAction(ctx context.Context, cmd *cli.Command) error {
	be, err := remote.NewBackendRemote(ctx, cmd, remote.BuckNaked())
	if err != nil {
		return err
	}

	client, err := be.Client()
	if err != nil {
		return err
	}

	runner := &QueryActionRunner[*tfe.Organization]{
		CommandName:  "oq",
		SchemaType:   reflect.TypeOf(tfe.Organization{}),
		DefaultAttrs: []string{"external-id:id", ".id:name"},
		FetchFn: func(ctx context.Context, cmd *cli.Command) (
			[]*tfe.Organization,
			error,
		) {
			options := tfe.OrganizationListOptions{
				ListOptions: tfe.ListOptions{
					PageNumber: 1,
					PageSize:   100,
				},
			}
			return PaginateWithOptions(
				ctx,
				cmd,
				&options,
				func(ctx context.Context, opts *tfe.OrganizationListOptions) (
					[]*tfe.Organization,
					*tfe.Pagination,
					error,
				) {
					page, err := client.Organizations.List(ctx, opts)
					if err != nil {
						return nil, nil, err
					}
					return page.Items, page.Pagination, nil
				},
				nil,
			)
		},
	}
	return runner.Run(ctx, cmd)
}

// oqCommandBuilder constructs the cli.Command for "oq", configuring metadata,
// flags, and the associated action/validator.
func oqCommandBuilder(meta meta.Meta) *cli.Command {
	return (&QueryCommandBuilder{
		Name:      "oq",
		Usage:     "organization query",
		UsageText: `tfctl oq [RootDir] [options]`,
		Flags: []cli.Flag{
			NewHostFlag("oq", meta.Config.Source),
		},
		Action: oqCommandAction,
		Meta:   meta,
	}).Build()
}
