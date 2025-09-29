// Copyright © 2025 Steve Taranto staranto@gmail.com
// SPDX-License-Identifier: MIT

package command

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"reflect"

	"github.com/hashicorp/jsonapi"
	"github.com/staranto/tfctlgo/internal/attrs"
	"github.com/staranto/tfctlgo/internal/meta"
	"github.com/staranto/tfctlgo/internal/output"
	"github.com/urfave/cli/v3"
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
	output.SliceDiceSpit(raw, al, cmd, "data", os.Stdout)
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
