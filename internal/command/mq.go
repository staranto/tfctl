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

// MqCommandAction is the action handler for the "mq" subcommand. It lists
// registry modules for the selected organization, supporting short-circuit
// behavior for --tldr and --schema, and emits results according to common
// output/attr flags.
func MqCommandAction(ctx context.Context, cmd *cli.Command) error {
	m := GetMeta(cmd)
	log.Debugf("Executing action for %v", m.Args[1:])

	// Bail out early if we're just dumping tldr.
	if ShortCircuitTLDR(ctx, cmd, "mq") {
		return nil
	}

	// Bail out early if we're just dumping the schema.
	if DumpSchemaIfRequested(cmd, reflect.TypeOf(tfe.RegistryModule{})) {
		return nil
	}

	attrs := BuildAttrs(cmd, ".id", "name")
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

	org, err := be.Organization()
	if err != nil {
		return fmt.Errorf("failed to resolve organization: %w", err)
	}

	options := tfe.RegistryModuleListOptions{
		ListOptions: tfe.ListOptions{PageNumber: 1, PageSize: 100},
	}

	var results []*tfe.RegistryModule

	// Paginate through the dataset
	for {
		page, err := client.RegistryModules.List(ctx, org, &options)
		if err != nil {
			ctxErr := remote.ErrorContext{
				Host:      be.Backend.Config.Hostname,
				Org:       org,
				Operation: "list registry modules",
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

// MqCommandBuilder constructs the cli.Command definition for the "mq" command,
// wiring flags, metadata, and the action/validator handlers.
func MqCommandBuilder(cmd *cli.Command, meta meta.Meta) *cli.Command {
	return &cli.Command{
		Name:      "mq",
		Usage:     "module registry query",
		UsageText: `tfctl mq [RootDir] [options]`,
		Metadata: map[string]any{
			"meta": meta,
		},
		Flags: append([]cli.Flag{
			NewHostFlag("mq", meta.Config.Source),
			NewOrgFlag("mq", meta.Config.Source),
			tldrFlag,
			schemaFlag,
		}, NewGlobalFlags("mq")...),
		Action: func(ctx context.Context, c *cli.Command) error {
			if err := MqCommandValidator(ctx, c); err != nil {
				return err
			}
			return MqCommandAction(ctx, c)
		},
	}
}

// MqCommandValidator performs command-level validation for "mq" and currently
// delegates to GlobalFlagsValidator for shared flag checks.
func MqCommandValidator(ctx context.Context, cmd *cli.Command) error {
	return GlobalFlagsValidator(ctx, cmd)
}
