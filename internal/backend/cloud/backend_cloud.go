// Copyright © 2025 Steve Taranto staranto@gmail.com
// SPDX-License-Identifier: MIT

package cloud

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/staranto/tfctlgo/internal/backend/remote"
	"github.com/urfave/cli/v3"
)

type BackendCloud struct {
	Ctx              context.Context
	Cmd              *cli.Command
	RootDir          string `json:"-" validate:"dir"`
	EnvOverride      string
	Version          int    `json:"version" validate:"gte=4"`
	TerraformVersion string `json:"terraform_version" validate:"semver"`
	Backend          struct {
		Type   string `json:"type"`
		Hash   int    `json:"hash"`
		Config struct {
			Hostname     string `json:"hostname" validate:"hostname"`
			Organization string `json:"organization" validate:"required"`
			Token        any    `json:"token"`
			Workspaces   struct {
				Name    string            `json:"name"`
				Project string            `json:"project"`
				Tags    map[string]string `json:"-"`
			} `json:"workspaces"`
		} `json:"config"`
	} `json:"backend"`
}

// Token retrieves the token from the environment variable, config file, or
// the credentials file, in that order.
func (be *BackendCloud) Token() (string, error) {
	var token string

	// Figure out if Token needs to be overridden by an environment variable.
	// The precedence is:
	// 1. TF_TOKEN_app_terraform_io
	// 2. TF_TOKEN
	// 3. Token in the config file
	// 4. Token in the TF credentials file.
	hostname := strings.ReplaceAll(be.Backend.Config.Hostname, ".", "_")
	if token = os.Getenv("TF_TOKEN_" + hostname); token == "" {
		token = os.Getenv("TF_TOKEN")
	}

	// If token was overridden by an environment variable, use that value and go
	// home early.
	if token != "" {
		return token, nil
	}

	token, _ = be.Backend.Config.Token.(string)

	// Once we're here, token may have existed already in the config file or it
	// may have been overridden by an environment variable.  If it's still empty,
	// we need to try to get it from the credentials file.
	if token == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}

		credsFile := home + "/.terraform.d/credentials.tfrc.json"
		data, err := os.ReadFile(credsFile)
		if err != nil {
			return "", fmt.Errorf("failed to read credentials file: %w", err)
		}

		var creds struct {
			Credentials map[string]struct {
				Token string `json:"token"`
			} `json:"credentials"`
		}

		if err := json.Unmarshal(data, &creds); err != nil {
			return "", fmt.Errorf("failed to unmarshal credentials file: %w", err)
		}

		if cred, ok := creds.Credentials[be.Backend.Config.Hostname]; ok {
			return cred.Token, nil
		}
	}

	return token, nil
}

func (c *BackendCloud) Transform2Remote(ctx context.Context, cmd *cli.Command) *remote.BackendRemote {
	remote := remote.BackendRemote{Ctx: ctx, Cmd: cmd}

	remote.RootDir = c.RootDir
	remote.Version = c.Version
	remote.TerraformVersion = c.TerraformVersion
	remote.EnvOverride = c.EnvOverride
	remote.Backend.Type = "remote"
	remote.Backend.Hash = c.Backend.Hash

	host := c.Backend.Config.Hostname
	if host == "" {
		host = cmd.String("host")
	}
	remote.Backend.Config.Hostname = host

	org := c.Backend.Config.Organization
	if org == "" {
		org = cmd.String("org")
	}
	remote.Backend.Config.Organization = org

	remote.Backend.Config.Workspaces.Name = c.Backend.Config.Workspaces.Name
	remote.Backend.Config.Token, _ = remote.Token()

	return &remote
}
