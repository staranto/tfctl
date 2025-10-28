// Copyright (c) 2025 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"context"
	"reflect"
	"strings"

	"github.com/apex/log"
	"github.com/hashicorp/go-tfe"
	altsrc "github.com/urfave/cli-altsrc/v3"
	yaml "github.com/urfave/cli-altsrc/v3/yaml"
	"github.com/urfave/cli/v3"

	"github.com/staranto/tfctlgo/internal/backend/remote"
	"github.com/staranto/tfctlgo/internal/filters"
	"github.com/staranto/tfctlgo/internal/meta"
)

// wqCommandAction is the action handler for the "wq" subcommand. It lists
// workspaces for the selected organization, supports --tldr/--schema short-
// circuits, and emits results per common flags.
func wqCommandAction(ctx context.Context, cmd *cli.Command) error {
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

			results, err := PaginateWithOptions(
				ctx,
				cmd,
				&options,
				func(ctx context.Context, opts *tfe.WorkspaceListOptions) (
					[]*tfe.Workspace,
					*tfe.Pagination,
					error,
				) {
					page, err := client.Workspaces.List(ctx, org, opts)
					if err != nil {
						ctxErr := OrgQueryErrorContext(
							be,
							org,
							"list workspaces",
						)
						return nil, nil, remote.FriendlyTFE(
							err,
							ctxErr,
						)
					}
					return page.Items, page.Pagination, nil
				},
				// Augmenter: customize options before each API call
				wqServerSideFilterAugmenter,
			)
			return results, err
		},
	}
	return runner.Run(ctx, cmd)
}

func wqServerSideFilterAugmenter(
	_ context.Context,
	cmd *cli.Command,
	opts *tfe.WorkspaceListOptions,
) error {
	spec := cmd.String("filter")
	filterList := filters.BuildFilters(spec)

	for _, f := range filterList {
		if f.ServerSide {
			parts := strings.Split(f.Key, ".")
			if len(parts) > 1 {
				switch parts[0] {
				case "project":
					opts.ProjectID = f.Value
				case "tag":
					opts.TagBindings = append(opts.TagBindings, &tfe.TagBinding{
						Key:   parts[1],
						Value: f.Value,
					})
				case "xtag":
					opts.ExcludeTags = parts[1]
				}
			}
		}
	}

	log.Debugf("opts after augmentation: %+v", opts)

	return nil
}

// wqCommandBuilder constructs the cli.Command for "wq", wiring metadata,
// flags, and action/validator handlers.
func wqCommandBuilder(meta meta.Meta) *cli.Command {
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
		Action: wqCommandAction,
		Meta:   meta,
	}).Build()
}
