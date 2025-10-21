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

// OqCommandAction is the action handler for the "oq" subcommand. It lists
// organizations from the configured host, supports --tldr/--schema
// short-circuit behavior, and emits output per common flags.
func OqCommandAction(ctx context.Context, cmd *cli.Command) error {
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
			return PaginateAndCollect(
				ctx,
				0,
				100,
				func(pageNumber, pageSize int) (
					[]*tfe.Organization,
					int,
					error,
				) {
					opts := tfe.OrganizationListOptions{
						ListOptions: tfe.ListOptions{
							PageNumber: pageNumber,
							PageSize:   pageSize,
						},
					}
					page, listErr := client.Organizations.List(
						ctx,
						&opts,
					)
					if listErr != nil {
						return nil, 0, listErr
					}
					return page.Items, page.Pagination.NextPage, nil
				},
			)
		},
	}
	return runner.Run(ctx, cmd)
}

// OqCommandBuilder constructs the cli.Command for "oq", configuring metadata,
// flags, and the associated action/validator.
func OqCommandBuilder(cmd *cli.Command, meta meta.Meta) *cli.Command {
	return (&QueryCommandBuilder{
		Name:      "oq",
		Usage:     "organization query",
		UsageText: `tfctl oq [RootDir] [options]`,
		Flags: []cli.Flag{
			NewHostFlag("oq", meta.Config.Source),
		},
		Action: OqCommandAction,
		Meta:   meta,
	}).Build()
}

// OqCommandValidator performs validation for "oq" and delegates shared checks
// to GlobalFlagsValidator.
func OqCommandValidator(ctx context.Context, cmd *cli.Command) error {
	return GlobalFlagsValidator(ctx, cmd)
}
