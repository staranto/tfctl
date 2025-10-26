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

// PqCommandAction is the action handler for the "pq" subcommand. It lists
// projects for the selected organization, supports --tldr/--schema
// short-circuit behavior, and emits output per common flags.
func PqCommandAction(ctx context.Context, cmd *cli.Command) error {
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
			return PaginateAndCollect(
				ctx,
				0,
				100,
				func(pageNumber, pageSize int) (
					[]*tfe.Project,
					int,
					error,
				) {
					opts := tfe.ProjectListOptions{
						ListOptions: tfe.ListOptions{
							PageNumber: pageNumber,
							PageSize:   pageSize,
						},
					}
					page, listErr := client.Projects.List(
						ctx,
						org,
						&opts,
					)
					if listErr != nil {
						ctxErr := OrgQueryErrorContext(
							be,
							org,
							"list projects",
						)
						return nil, 0, remote.FriendlyTFE(
							listErr,
							ctxErr,
						)
					}
					return page.Items, page.Pagination.NextPage, nil
				},
			)
		},
	}
	return runner.Run(ctx, cmd)
}

// PqCommandBuilder constructs the cli.Command for "pq", wiring metadata,
// flags, and action/validator handlers.
func PqCommandBuilder(cmd *cli.Command, meta meta.Meta) *cli.Command {
	return (&QueryCommandBuilder{
		Name:  "pq",
		Usage: "project query",
		Flags: []cli.Flag{
			NewHostFlag("pq", meta.Config.Source),
			NewOrgFlag("pq", meta.Config.Source),
		},
		Action: PqCommandAction,
		Meta:   meta,
	}).Build()
}
