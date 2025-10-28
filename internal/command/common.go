// Copyright (c) 2025 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"reflect"

	"github.com/apex/log"
	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/jsonapi"
	"github.com/urfave/cli/v3"

	"github.com/staranto/tfctlgo/internal/attrs"
	"github.com/staranto/tfctlgo/internal/backend"
	"github.com/staranto/tfctlgo/internal/backend/remote"
	"github.com/staranto/tfctlgo/internal/meta"
	"github.com/staranto/tfctlgo/internal/output"
)

// ShortCircuitTLDR checks the --tldr flag and, if present and available,
// runs `tldr tfctl <subcmd>` and returns true so the caller can exit early.
func ShortCircuitTLDR(ctx context.Context, cmd *cli.Command, subcmd string) bool {
	if cmd.Bool("tldr") {
		if _, err := exec.LookPath("tldr"); err == nil {
			c := exec.CommandContext(ctx, "tldr", "tfctl", subcmd)
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			_ = c.Run()
		}
		return true
	}
	return false
}

// DumpSchemaIfRequested prints the JSON schema for the provided type when
// --schema is set, and returns true if it handled the request.
func DumpSchemaIfRequested(cmd *cli.Command, t reflect.Type) bool {
	if cmd.Bool("schema") {
		output.DumpSchema("", t)
		return true
	}
	return false
}

// BuildAttrs constructs an AttrList with defaults and optional extras from
// --attrs, then applies the global transform spec.
func BuildAttrs(cmd *cli.Command, defaults ...string) (al attrs.AttrList) {
	//nolint:errcheck
	{
		for _, d := range defaults {
			al.Set(d)
		}
		if extras := cmd.String("attrs"); extras != "" {
			al.Set(extras)
		}
		al.SetGlobalTransformSpec()
	}
	return
}

// EmitJSONAPISlice marshals a slice as JSONAPI and passes it to the common
// output routine.
func EmitJSONAPISlice(results any, al attrs.AttrList, cmd *cli.Command) error {
	var raw bytes.Buffer
	if err := jsonapi.MarshalPayload(&raw, results); err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	output.SliceDiceSpit(raw, al, cmd, "data", os.Stdout, nil)
	return nil
}

// GetMeta returns the meta.Meta stored in the command's Metadata. If missing
// or of an unexpected type, it returns the zero value.
func GetMeta(cmd *cli.Command) meta.Meta {
	if cmd == nil || cmd.Metadata == nil {
		return meta.Meta{}
	}
	if m, ok := cmd.Metadata["meta"].(meta.Meta); ok {
		return m
	}
	return meta.Meta{}
}

// Augmenter[O] is a callback function that customizes options before
// each API call. It receives the context, command, and a pointer to the
// options object, allowing mutation of options based on command flags or
// other context. Return an error to abort pagination.
type Augmenter[O any] func(
	context.Context,
	*cli.Command,
	*O,
) error

// PaginateWithOptions[T, O] is a generic paginator that drives paginated API
// calls with mutable options. It handles pagination logic and returns all
// collected results. The augmenter callback (if provided) is called before
// each API invocation, allowing options customization (e.g., setting filters
// or tags). The fetcher callback encapsulates the actual API call and must
// return results, pagination info, and any error.
func PaginateWithOptions[T, O any](
	ctx context.Context,
	cmd *cli.Command,
	options *O,
	fetcher func(context.Context, *O) ([]T, *tfe.Pagination, error),
	augmenter Augmenter[O],
) ([]T, error) {
	var results []T

	// Paginate through pages
	for {
		// Invoke augmenter before each page (to allow options mutation)
		if augmenter != nil {
			if err := augmenter(ctx, cmd, options); err != nil {
				return nil, err
			}
		}

		// Fetch current page
		items, pagination, err := fetcher(ctx, options)
		if err != nil {
			return nil, err
		}

		results = append(results, items...)

		// Check if there are more pages
		if pagination.NextPage == 0 {
			break
		}

		// Increment page number for next iteration
		setPageNumber(options, pagination.NextPage)
	}

	return results, nil
}

// setPageNumber uses reflection to set the PageNumber field in the options
// struct. It assumes the struct has a ListOptions.PageNumber field (standard
// in tfe API options).
func setPageNumber(options any, pageNumber int) {
	v := reflect.ValueOf(options).Elem()
	lo := v.FieldByName("ListOptions")
	if !lo.IsValid() {
		return
	}
	pn := lo.FieldByName("PageNumber")
	if pn.IsValid() && pn.CanSet() {
		pn.SetInt(int64(pageNumber))
	}
}

// OrgQueryErrorContext is a helper to construct remote.ErrorContext for
// organization-related queries (mq, pq). It requires the backend and
// organization.
func OrgQueryErrorContext(
	be *remote.BackendRemote,
	org string,
	operation string,
) remote.ErrorContext {
	return remote.ErrorContext{
		Host:      be.Backend.Config.Hostname,
		Org:       org,
		Operation: operation,
		Resource:  "organization",
	}
}

// QueryCommandBuilder is a helper that constructs a cli.Command for query
// subcommands (mq, pq, oq, svq, rq, wq) using a consistent pattern.
// It accepts the command name, usage text, optional UsageText, custom flags,
// the action handler, and meta. The builder automatically wires metadata,
// adds tldr/schema flags, applies global flags, and sets up validators.
type QueryCommandBuilder struct {
	Name      string
	Usage     string
	UsageText string
	Flags     []cli.Flag
	Action    func(context.Context, *cli.Command) error
	Meta      meta.Meta
}

// Build returns a configured cli.Command from the builder.
func (qcb *QueryCommandBuilder) Build() *cli.Command {
	return &cli.Command{
		Name:      qcb.Name,
		Usage:     qcb.Usage,
		UsageText: qcb.UsageText,
		Metadata: map[string]any{
			"meta": qcb.Meta,
		},
		Flags: append(qcb.Flags, append([]cli.Flag{
			tldrFlag,
			schemaFlag,
		}, NewGlobalFlags(qcb.Name)...)...),
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			return ctx, GlobalFlagsValidator(ctx, c)
		},
		Action: qcb.Action,
	}
}

// QueryActionRunner[T] encapsulates the common query action pattern for all
// query subcommands. It handles steps 1-4 and 6 (GetMeta, short-circuit
// checks, BuildAttrs, schema dumping, and output emission), with step 5
// (data fetching) provided by FetchFn.
type QueryActionRunner[T any] struct {
	CommandName  string
	SchemaType   reflect.Type
	DefaultAttrs []string
	FetchFn      func(context.Context, *cli.Command) ([]T, error)
}

// Run executes the query action with the provided context and command.
func (qar *QueryActionRunner[T]) Run(
	ctx context.Context,
	cmd *cli.Command,
) error {
	// Step 1: GetMeta + debug.
	m := GetMeta(cmd)
	log.Debugf("Executing action for %v", m.Args[1:])

	// Step 2: Short-circuit checks.
	if ShortCircuitTLDR(ctx, cmd, qar.CommandName) {
		return nil
	}
	if DumpSchemaIfRequested(cmd, qar.SchemaType) {
		return nil
	}

	// Step 3: BuildAttrs + debug.
	attrs := BuildAttrs(cmd, qar.DefaultAttrs...)
	log.Debugf("attrs: %v", attrs)

	// Step 4: Fetch data.
	results, err := qar.FetchFn(ctx, cmd)
	if err != nil {
		return err
	}

	// Step 5: Emit + return.
	if err := EmitJSONAPISlice(results, attrs, cmd); err != nil {
		return err
	}
	return nil
}

// InitRemoteOrgQuery initializes a remote backend connection for queries that
// operate exclusively on organizations. It returns the backend, organization
// name, and TFE client, or an error if initialization fails.
func InitRemoteOrgQuery(
	ctx context.Context,
	cmd *cli.Command,
) (*remote.BackendRemote, string, *tfe.Client, error) {
	be, err := remote.NewBackendRemote(ctx, cmd, remote.BuckNaked())
	if err != nil {
		return nil, "", nil, err
	}
	log.Debugf("be: %v", be)

	client, err := be.Client()
	if err != nil {
		return nil, "", nil, err
	}
	log.Debugf("client: %v", client.BaseURL())

	org, err := be.Organization()
	if err != nil {
		return nil, "", nil, fmt.Errorf(
			"failed to resolve organization: %w",
			err,
		)
	}

	return be, org, client, nil
}

// InitLocalBackendQuery initializes a local backend connection for queries
// that operate on local state. It returns the backend or an error if
// initialization fails.
func InitLocalBackendQuery(ctx context.Context, cmd *cli.Command) (
	backend.Backend,
	error,
) {
	be, err := backend.NewBackend(ctx, *cmd)
	if err != nil {
		return nil, err
	}
	log.Debugf("be: %v", be)
	return be, nil
}
