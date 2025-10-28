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

// pqCommandAction is the action handler for the "pq" subcommand. It lists
// projects for the selected organization, supports --tldr/--schema
// short-circuit behavior, and emits output per common flags.
func pqCommandAction(ctx context.Context, cmd *cli.Command) error {
	be, org, client, err := InitRemoteOrgQuery(ctx, cmd)
	if err != nil {
		return err
	}

	runner := &QueryActionRunner[*tfe.Project]{
		CommandName:  "pq",
		SchemaType:   reflect.TypeOf(tfe.Project{}),
		DefaultAttrs: []string{".id", "name"},
		FetchFn: func(ctx context.Context, cmd *cli.Command) (
			[]*tfe.Project,
			error,
		) {
			options := tfe.ProjectListOptions{
				ListOptions: tfe.ListOptions{
					PageNumber: 1,
					PageSize:   100,
				},
			}
			return PaginateWithOptions(
				ctx,
				cmd,
				&options,
				func(ctx context.Context, opts *tfe.ProjectListOptions) (
					[]*tfe.Project,
					*tfe.Pagination,
					error,
				) {
					page, err := client.Projects.List(ctx, org, opts)
					if err != nil {
						ctxErr := OrgQueryErrorContext(
							be,
							org,
							"list projects",
						)
						return nil, nil, remote.FriendlyTFE(
							err,
							ctxErr,
						)
					}
					return page.Items, page.Pagination, nil
				},
				nil,
			)
		},
	}
	return runner.Run(ctx, cmd)
}

// pqCommandBuilder constructs the cli.Command for "pq", wiring metadata,
// flags, and action/validator handlers.
func pqCommandBuilder(meta meta.Meta) *cli.Command {
	return (&QueryCommandBuilder{
		Name:  "pq",
		Usage: "project query",
		Flags: []cli.Flag{
			NewHostFlag("pq", meta.Config.Source),
			NewOrgFlag("pq", meta.Config.Source),
		},
		Action: pqCommandAction,
		Meta:   meta,
	}).Build()
}
