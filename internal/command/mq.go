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

// MqCommandAction is the action handler for the "mq" subcommand. It lists
// registry modules for the selected organization, supporting short-circuit
// behavior for --tldr and --schema, and emits results according to common
// output/attr flags.
func MqCommandAction(ctx context.Context, cmd *cli.Command) error {
	be, org, client, err := InitRemoteOrgQuery(ctx, cmd)
	if err != nil {
		return err
	}

	runner := &QueryActionRunner[*tfe.RegistryModule]{
		CommandName:  "mq",
		SchemaType:   reflect.TypeOf(tfe.RegistryModule{}),
		DefaultAttrs: []string{".id", "name"},
		FetchFn: func(ctx context.Context, cmd *cli.Command) (
			[]*tfe.RegistryModule,
			error,
		) {
			return PaginateAndCollect(
				ctx,
				0,
				100,
				func(pageNumber, pageSize int) (
					[]*tfe.RegistryModule,
					int,
					error,
				) {
					opts := tfe.RegistryModuleListOptions{
						ListOptions: tfe.ListOptions{
							PageNumber: pageNumber,
							PageSize:   pageSize,
						},
					}
					page, listErr := client.RegistryModules.List(
						ctx,
						org,
						&opts,
					)
					if listErr != nil {
						ctxErr := OrgQueryErrorContext(
							be,
							org,
							"list registry modules",
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

// MqCommandBuilder constructs the cli.Command definition for the "mq" command,
// wiring flags, metadata, and the action/validator handlers.
func MqCommandBuilder(cmd *cli.Command, meta meta.Meta) *cli.Command {
	return (&QueryCommandBuilder{
		Name:      "mq",
		Usage:     "module registry query",
		UsageText: `tfctl mq [RootDir] [options]`,
		Flags: []cli.Flag{
			NewHostFlag("mq", meta.Config.Source),
			NewOrgFlag("mq", meta.Config.Source),
		},
		Action: MqCommandAction,
		Meta:   meta,
	}).Build()
}
