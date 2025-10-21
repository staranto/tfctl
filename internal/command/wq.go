// Copyright (c) 2025 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"context"
	"reflect"

	"github.com/hashicorp/go-tfe"
	altsrc "github.com/urfave/cli-altsrc/v3"
	yaml "github.com/urfave/cli-altsrc/v3/yaml"
	"github.com/urfave/cli/v3"

	"github.com/staranto/tfctlgo/internal/backend/remote"
	"github.com/staranto/tfctlgo/internal/meta"
)

// WqCommandAction is the action handler for the "wq" subcommand. It lists
// workspaces for the selected organization, supports --tldr/--schema short-
// circuits, and emits results per common flags.
func WqCommandAction(ctx context.Context, cmd *cli.Command) error {
	be, org, client, err := InitRemoteOrgQuery(ctx, cmd)
	if err != nil {
		return err
	}

	runner := &QueryActionRunner[*tfe.Workspace]{
		CommandName:  "wq",
		SchemaType:   reflect.TypeOf(tfe.Workspace{}),
		DefaultAttrs: []string{".id", "name"},
		FetchFn: func(ctx context.Context, cmd *cli.Command) (
			[]*tfe.Workspace,
			error,
		) {
			options := tfe.WorkspaceListOptions{
				ListOptions: tfe.ListOptions{
					PageNumber: 1,
					PageSize:   100,
				},
			}

			var results []*tfe.Workspace

			// Paginate through the dataset
			for {
				page, err := client.Workspaces.List(ctx, org, &options)
				if err != nil {
					ctxErr := OrgQueryErrorContext(
						be,
						org,
						"list workspaces",
					)
					return nil, remote.FriendlyTFE(err, ctxErr)
				}

				results = append(results, page.Items...)

				if page.Pagination.NextPage == 0 {
					break
				}
				options.ListOptions.PageNumber++
			}

			return results, nil
		},
	}
	return runner.Run(ctx, cmd)
}

// WqCommandBuilder constructs the cli.Command for "wq", wiring metadata,
// flags, and action/validator handlers.
func WqCommandBuilder(cmd *cli.Command, meta meta.Meta) *cli.Command {
	return (&QueryCommandBuilder{
		Name:      "wq",
		Usage:     "workspace query",
		UsageText: `tfctl wq [RootDir] [options]`,
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "limit",
				Aliases: []string{"l"},
				Usage:   "limit workspaces returned",
				Sources: cli.NewValueSourceChain(
					yaml.YAML("wq.limit", altsrc.StringSourcer(meta.Config.Source)),
					yaml.YAML("limit", altsrc.StringSourcer(meta.Config.Source)),
				),
				Value: 99999,
			},
			NewHostFlag("wq", meta.Config.Source),
			NewOrgFlag("wq", meta.Config.Source),
		},
		Action: WqCommandAction,
		Meta:   meta,
	}).Build()
}

// WqCommandValidator performs validation for "wq" and delegates to
// GlobalFlagsValidator.
func WqCommandValidator(ctx context.Context, cmd *cli.Command) error {
	return GlobalFlagsValidator(ctx, cmd)
}
