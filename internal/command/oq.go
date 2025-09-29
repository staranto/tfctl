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

// OqCommandAction is the action handler for the "oq" subcommand. It lists
// organizations from the configured host, supports --tldr/--schema
// short-circuit behavior, and emits output per common flags.
func OqCommandAction(ctx context.Context, cmd *cli.Command) error {
	m := GetMeta(cmd)
	log.Debugf("Executing action for %v", m.Args[1:])

	// Bail out early if we're just dumping tldr.
	if ShortCircuitTLDR(ctx, cmd, "oq") {
		return nil
	}

	//

	// Bail out early if we're just dumping the schema.
	if DumpSchemaIfRequested(cmd, reflect.TypeOf(tfe.Organization{})) {
		return nil
	}

	attrs := BuildAttrs(cmd, "external-id:id", ".id:name")
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

	options := tfe.OrganizationListOptions{
		ListOptions: tfe.ListOptions{PageNumber: 1, PageSize: 100},
	}

	var results []*tfe.Organization

	// Paginate through the dataset
	for {
		page, err := client.Organizations.List(ctx, &options)
		if err != nil {
			return fmt.Errorf("failed to list organizations: %w", err)
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

// OqCommandBuilder constructs the cli.Command for "oq", configuring metadata,
// flags, and the associated action/validator.
func OqCommandBuilder(cmd *cli.Command, meta meta.Meta) *cli.Command {
	return &cli.Command{
		Name:      "oq",
		Usage:     "organization query",
		UsageText: `tfctl oq [RootDir] [options]`,
		Metadata: map[string]any{
			"meta": meta,
		},
		Flags: append([]cli.Flag{
			tldrFlag,
			NewHostFlag("oq", meta.Config.Source),
			schemaFlag,
		}, NewGlobalFlags("oq")...),
		Action: func(ctx context.Context, c *cli.Command) error {
			if err := OqCommandValidator(ctx, c); err != nil {
				return err
			}
			return OqCommandAction(ctx, c)
		},
	}
}

// OqCommandValidator performs validation for "oq" and delegates shared checks
// to GlobalFlagsValidator.
func OqCommandValidator(ctx context.Context, cmd *cli.Command) error {
	return GlobalFlagsValidator(ctx, cmd)
}
