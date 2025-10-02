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

// PqCommandAction is the action handler for the "pq" subcommand. It lists
// projects for the selected organization, supports --tldr/--schema
// short-circuit behavior, and emits output per common flags.
func PqCommandAction(ctx context.Context, cmd *cli.Command) error {
	m := GetMeta(cmd)
	log.Debugf("Executing action for %v", m.Args[1:])

	// Bail out early if we're just dumping tldr.
	if ShortCircuitTLDR(ctx, cmd, "pq") {
		return nil
	}

	//

	// Bail out early if we're just dumping the schema.
	if DumpSchemaIfRequested(cmd, reflect.TypeOf(tfe.Project{})) {
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

	org, err := be.Organization()
	if err != nil {
		return fmt.Errorf("failed to resolve organization: %w", err)
	}

	options := tfe.ProjectListOptions{
		ListOptions: tfe.ListOptions{PageNumber: 1, PageSize: 100},
	}

	var results []*tfe.Project

	// Paginate through the dataset
	for {
		page, err := client.Projects.List(ctx, org, &options)
		if err != nil {
			ctxErr := remote.ErrorContext{
				Host:      be.Backend.Config.Hostname,
				Org:       org,
				Operation: "list projects",
				Resource:  "organization",
			}
			return remote.FriendlyTFE(err, ctxErr)
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

// PqCommandBuilder constructs the cli.Command for "pq", wiring metadata,
// flags, and action/validator handlers.
func PqCommandBuilder(cmd *cli.Command, meta meta.Meta) *cli.Command {
	return &cli.Command{
		Name:  "pq",
		Usage: "project query",
		Metadata: map[string]any{
			"meta": meta,
		},
		Flags: append([]cli.Flag{
			NewHostFlag("pq", meta.Config.Source),
			NewOrgFlag("pq", meta.Config.Source),
			tldrFlag,
			schemaFlag,
		}, NewGlobalFlags("pq")...),
		Action: func(ctx context.Context, c *cli.Command) error {
			if err := PqCommandValidator(ctx, c); err != nil {
				return err
			}
			return PqCommandAction(ctx, c)
		},
	}
}

// PqCommandValidator performs validation for "pq" and delegates shared checks
// to GlobalFlagsValidator.
func PqCommandValidator(ctx context.Context, cmd *cli.Command) error {
	return GlobalFlagsValidator(ctx, cmd)
}
