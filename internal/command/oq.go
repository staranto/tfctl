// Copyright © 2025 Steve Taranto staranto@gmail.com
// SPDX-License-Identifier: MIT

package command

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"reflect"

	"github.com/apex/log"
	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/jsonapi"
	"github.com/staranto/tfctlgo/internal/attrs"
	"github.com/staranto/tfctlgo/internal/backend/remote"
	"github.com/staranto/tfctlgo/internal/meta"
	"github.com/staranto/tfctlgo/internal/output"
	"github.com/urfave/cli/v3"
)

func OqCommandAction(ctx context.Context, cmd *cli.Command) error {
	meta := cmd.Metadata["meta"].(meta.Meta)
	log.Debugf("Executing action for %v", meta.Args[1:])

	// Bail out early if we're just dumping examples.
	if cmd.Bool("examples") {
		examples := [][2]string{
			{"tfctl oq", "All Organizations on the default HCP/TFE server."},
			{"tfctl oq --host tfe.example.com", "All Organizations on the tfe.example.com server."},
			{"tfctl oq --attrs email", "All Organizations, plus the email of their administrator."},
			{"tfctl oq --filter 'name@prod'", "Organizations containing 'prod' in their name."},
		}
		output.DumpExamples(ctx, cmd, examples)
		return nil
	}

	// Bail out early if we're just dumping the schema.
	if cmd.Bool("schema") {
		output.DumpSchema("", reflect.TypeOf(tfe.Organization{}))
		return nil
	}

	var attrs attrs.AttrList
	//nolint:errcheck
	{
		attrs.Set("external-id:id")
		attrs.Set(".id:name")

		extras := cmd.String("attrs")
		if extras != "" {
			attrs.Set(extras)
		}

		attrs.SetGlobalTransformSpec()
	}
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

	// Marshal into a JSON document so we can slice and dice some more.  Note that
	// we're using jsonapi, which will use the StructField tags as the keys of the
	// JSON document.
	var raw bytes.Buffer
	if err := jsonapi.MarshalPayload(&raw, results); err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	output.SliceDiceSpit(raw, attrs, cmd, "data", os.Stdout)

	return nil
}

func OqCommandBuilder(cmd *cli.Command, meta meta.Meta, globalFlags []cli.Flag) *cli.Command {
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
		}, globalFlags...),
		Action: func(ctx context.Context, c *cli.Command) error {
			if err := OqCommandValidator(ctx, c, globalFlags); err != nil {
				return err
			}
			return OqCommandAction(ctx, c)
		},
	}
}

func OqCommandValidator(ctx context.Context, cmd *cli.Command, globalFlags []cli.Flag) error {
	return GlobalFlagsValidator(ctx, cmd, globalFlags)
}
