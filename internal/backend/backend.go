// Copyright © 2025 Steve Taranto staranto@gmail.com
// SPDX-License-Identifier: MIT

package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/hashicorp/go-tfe"
	"github.com/staranto/tfctlgo/internal/backend/cloud"
	"github.com/staranto/tfctlgo/internal/backend/local"
	"github.com/staranto/tfctlgo/internal/backend/remote"
	"github.com/staranto/tfctlgo/internal/backend/s3"
	"github.com/staranto/tfctlgo/internal/meta"
	"github.com/urfave/cli/v3"
)

type BackendType struct {
	Ctx         context.Context
	Cmd         *cli.Command
	RootDir     string `json:"-" validate:"dir"`
	EnvOverride string
	// Version          int    `json:"version" validate:"gte=4"`
	// TerraformVersion string `json:"terraform_version" validate:"semver"`
}

type Backend interface {
	Runs() ([]*tfe.Run, error)
	// State() returns the CSV~0 state document.
	State() ([]byte, error)
	// States() returns the state documents specified by the specs.
	States(...string) ([][]byte, error)
	StateVersions() ([]*tfe.StateVersion, error)
	String() string
	Type() (string, error)
}

type SelfDiffer interface {
	DiffStates(ctx context.Context, cmd *cli.Command) ([][]byte, error)
}

func NewBackend(ctx context.Context, cmd cli.Command) (Backend, error) {
	meta := cmd.Metadata["meta"].(meta.Meta)
	log.Debugf("NewBackend: meta: %v", meta)

	cFile, cErr := os.Stat(filepath.Join(meta.RootDir, ".terraform", "terraform.tfstate"))
	sFile, sErr := os.Stat(filepath.Join(meta.RootDir, "terraform.tfstate"))
	eFile, eErr := os.Stat(filepath.Join(meta.RootDir, ".terraform", "environment"))
	_, _, _ = cFile, sFile, eFile // HACK

	// Maybe we're in a non-sq command and just need a naked remote.  This will be
	// when c, s and e are all in error meaning none of them exist.
	if cErr != nil && sErr != nil && eErr != nil {
		return remote.NewBackendRemote(ctx, &cmd, remote.BuckNaked())
	}

	// If terraform.tfstate exists but .terraform/terraform.tfstate doesn't,
	// infer local backend.  This is a terraform.backend {} block situation
	if cErr != nil && sErr == nil {
		return local.NewBackendLocal(ctx, &cmd,
			local.FromRootDir(meta.RootDir),
			local.WithEnvOverride(meta.Env),
		)
	}

	// Peek at the backend type so we can switch on it.
	// TODO We're double reading the file.  Once in peek() and once in the New().
	typ, err := peek(meta)
	if err != nil {
		return nil, err
	}

	switch typ {
	case "cloud":
		be, err := cloud.NewBackendCloud(ctx, &cmd,
			cloud.FromRootDir(meta.RootDir),
			cloud.WithEnvOverride(meta.Env),
		)
		return be.Transform2Remote(ctx, &cmd), err
	case "local":
		be, err := local.NewBackendLocal(ctx, &cmd,
			local.FromRootDir(meta.RootDir),
			local.WithEnvOverride(meta.Env),
		)
		return be, err
	case "remote":
		be, err := remote.NewBackendRemote(ctx, &cmd,
			remote.FromRootDir(meta.RootDir),
			remote.WithEnvOverride(meta.Env),
			remote.WithSvOverride(),
		)
		return be, err
	case "s3":
		be, err := s3.NewBackendS3(ctx, &cmd,
			s3.FromRootDir(meta.RootDir),
			s3.WithEnvOverride(meta.Env),
			s3.WithSvOverride(),
		)
		return be, err
	}

	// This is a fail-safe.  We should never get here.
	return nil, fmt.Errorf("unknown type %s: %w", typ, err)
}

func peek(meta meta.Meta) (string, error) {
	raw, err := os.ReadFile(filepath.Join(meta.RootDir, ".terraform", "terraform.tfstate"))
	if err != nil {
		return "", err
	}

	var peeker map[string]json.RawMessage
	if err := json.Unmarshal(raw, &peeker); err != nil {
		return "", fmt.Errorf("Can't peek: %w", err)
	}

	if err := json.Unmarshal(peeker["backend"], &peeker); err != nil {
		return "", fmt.Errorf("Can't peek: %w", err)
	}

	var typ string
	if err := json.Unmarshal(peeker["type"], &typ); err != nil {
		return "", fmt.Errorf("Can't peek: %w", err)
	}
	log.Debugf("type: %s", typ)

	return typ, nil
}
