// Copyright (c) 2025 Steve Taranto staranto@gmail.com.
// SPDX-License-Identifier: Apache-2.0

package remote

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/apex/log"
	"github.com/hashicorp/go-tfe"
	"github.com/urfave/cli/v3"

	"github.com/staranto/tfctlgo/internal/csv"
	"github.com/staranto/tfctlgo/internal/differ"
)

type BackendRemote struct {
	Ctx              context.Context
	Cmd              *cli.Command
	RootDir          string `json:"-" validate:"dir"`
	EnvOverride      string
	SvOverride       string
	RunList          []*tfe.Run
	StateVersionList []*tfe.StateVersion
	Version          int    `json:"version" validate:"gte=4"`
	TerraformVersion string `json:"terraform_version" validate:"semver"`
	Backend          struct {
		Type   string `json:"type" validate:"eq=remote"`
		Hash   int    `json:"hash"`
		Config struct {
			Hostname     string `json:"hostname" validate:"hostname"`
			Organization string `json:"organization" validate:"required"`
			Token        any    `json:"token"`
			Workspaces   struct {
				Name   string `json:"name" validate:"required_without=Prefix"`
				Prefix string `json:"prefix" validate:"required_without=Name"`
			} `json:"workspaces"`
		} `json:"config"`
	} `json:"backend"`
}

// Sentinel errors for validation and unsupported cases. These enable callers to
// detect specific conditions via errors.Is/As while keeping messages consistent.
var (
	ErrTokenNotString                = errors.New("token is not a string")
	ErrInvalidClientType             = errors.New("not a Cloud or Enterprise TFE server")
	ErrOrganizationNotSet            = errors.New("organization is not set")
	ErrNoCurrentStateVersion         = errors.New("no current state version")
	ErrURLNotSupported               = errors.New("URL not supported")
	ErrWorkspaceNameAndPrefixBothSet = errors.New("both workspace name and prefix are set")
)

// Client optionally validates and returns a TFE client to the host specified
// in the remote backend.
func (be *BackendRemote) Client(validate ...bool) (*tfe.Client, error) {
	beCfg := be.Backend.Config

	// Resolve token using standard precedence (env, config, credentials file).
	token, err := be.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve token: %w", err)
	}

	client, err := tfe.NewClient(&tfe.Config{
		Address: "https://" + beCfg.Hostname,
		Token:   token,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create TFE client: %w", err)
	}

	if len(validate) > 0 && validate[0] {
		if !(client.IsCloud() || client.IsEnterprise()) {
			return nil, fmt.Errorf("failed to validate TFE client: %w", ErrInvalidClientType)
		}
	}

	return client, nil
}

func (be *BackendRemote) Organization() (string, error) {
	org := be.Cmd.String("org")
	if org != "" {
		return org, nil
	}

	org = be.Backend.Config.Organization
	if org == "" {
		return "", fmt.Errorf("organization is not set (precedence: --org > backend block > tfctl.yaml). Set --org or backend.config.organization: %w", ErrOrganizationNotSet)
	}

	return org, nil
}

func (be *BackendRemote) DiffStates(ctx context.Context, cmd *cli.Command) ([][]byte, error) {
	// Fixup diffArgs
	svSpecs := []string{"CSV~1", "CSV~0"}

	diffArgs := differ.ParseDiffArgs(ctx, cmd)

	switch len(diffArgs) {
	case 0:
		// No args, so use the last two states.
	case 1:
		if strings.HasPrefix(diffArgs[0], "+") {
			// limit := 9999
			// if l, err := strconv.Atoi(diffArgs[0][1:]); err == nil {
			// 	limit = l
			// }

			var err error
			be.StateVersionList, err = be.StateVersions( /* TODO limit */ )
			if err != nil {
				return nil, err
			}

			selectedVersions := differ.SelectStateVersions(be.StateVersionList)

			log.Debugf("selectedVersions: %d", len(selectedVersions))

			if len(selectedVersions) == 0 {
				return nil, nil
			} else if len(selectedVersions) == 2 {
				svSpecs[0] = selectedVersions[1].ID
				svSpecs[1] = selectedVersions[0].ID
			}
		} else {
			svSpecs[0] = diffArgs[0]
		}
	case 2:
		svSpecs = diffArgs
	}

	states, err := be.States(svSpecs[0], svSpecs[1])
	if err != nil {
		return nil, fmt.Errorf("failed to get states: %w", err)
	}

	return states, nil
}

func (be *BackendRemote) State() ([]byte, error) {
	sv := be.Cmd.String("sv")
	states, err := be.States(sv)
	if err != nil {
		return nil, err
	}
	return states[0], nil
}

func (be *BackendRemote) States(specs ...string) ([][]byte, error) {
	var results [][]byte

	candidates, err := be.StateVersions()
	if err != nil {
		return nil, err
	}
	versions, err := csv.Finder(candidates, specs...)
	if err != nil {
		return nil, err
	}
	log.Debugf("versions: %v", versions)

	// Now pound through the found versions and return each of their state bodies.
	for _, v := range versions {
		doc, err := Hitter(be, v.DownloadURL)
		if err != nil {
			return nil, fmt.Errorf("failed to get state: %w", err)
		}
		results = append(results, doc.Bytes())
	}

	return results, nil
}

func (be *BackendRemote) StateVersion(svSpecs ...string) (tfe.StateVersion, error) {
	if len(svSpecs) == 0 {
		svSpecs = []string{"CSV~0"}
	}

	// We just want to fall through if svSpec[0] is an sv-id, so don't bother
	// checking for it.

	// Force CSV~ prefix to uppercase to avoid case sensitivity issues
	if strings.HasPrefix(strings.ToUpper(svSpecs[0]), "CSV~") {
		svSpecs[0] = strings.ToUpper(svSpecs[0])
	}

	// If no svid was passed in or it's CSV~0, we'll short circuit this and try to
	// get the current state version.
	if svSpecs[0] == "" || svSpecs[0] == "CSV~0" {
		workspace, err := be.Workspace()
		if err != nil {
			return tfe.StateVersion{}, fmt.Errorf("failed to get workspace: %w", err)
		}

		if workspace.CurrentStateVersion == nil {
			return tfe.StateVersion{},
				fmt.Errorf("workspace %s has no current state version: %w",
					workspace.ID, ErrNoCurrentStateVersion)
		}
		svSpecs[0] = workspace.CurrentStateVersion.ID
	} else if strings.HasPrefix(svSpecs[0], "CSV~") {
		// We've got to search through the state versions to be able to grab the
		// relative one.
		if be.StateVersionList == nil {
			be.StateVersionList, _ = be.StateVersions()
		}

		parts := strings.Split(svSpecs[0], "~")
		offset, err := strconv.Atoi(parts[1])
		if err != nil {
			return tfe.StateVersion{}, fmt.Errorf("invalid state version offset: %w", err)
		}

		svSpecs[0] = be.StateVersionList[offset].ID
	} else if serial, err := strconv.ParseInt(svSpecs[0], 10, 64); err == nil {
		// If we've got an int, find that specific serial number.
		if be.StateVersionList == nil {
			be.StateVersionList, _ = be.StateVersions()
		}

		for _, sv := range be.StateVersionList {
			if sv.Serial == serial {
				svSpecs[0] = sv.ID
				break
			}
		}
	} else if strings.HasPrefix(svSpecs[0], "https://") {
		return tfe.StateVersion{}, fmt.Errorf("URL not supported: %w", ErrURLNotSupported)
	}

	// First look to see if it's in the cache.  If it is, unmarshall it and return
	// early.
	if entry, ok := CacheReader(be, svSpecs[0]); ok {
		var stateVersion tfe.StateVersion
		if err := json.Unmarshal(entry.Data, &stateVersion); err != nil {
			return tfe.StateVersion{}, fmt.Errorf("failed to unmarshal state version: %w", err)
		}
		return stateVersion, nil
	}

	client, err := be.Client()
	if err != nil {
		return tfe.StateVersion{}, fmt.Errorf("failed to get TFE client: %w", err)
	}
	ctx := context.Background()

	stateVersion, err := client.StateVersions.Read(ctx, svSpecs[0])
	if err != nil {
		return tfe.StateVersion{}, fmt.Errorf("failed to get state version: %w", err)
	}

	// If we got here, we need to write the state version to the cache.
	stateVersionBytes, err := json.Marshal(stateVersion)
	if err == nil {
		if err := CacheWriter(be, svSpecs[0], stateVersionBytes); err != nil {
			log.WithError(err).Warn("failed to write state version to cache")
		}
	}

	return *stateVersion, nil
}

func (be *BackendRemote) Runs() ([]*tfe.Run, error) {
	if len(be.RunList) > 0 {
		log.Errorf("be.RunList: preloaded with %d", len(be.RunList))
		return be.RunList, nil
	}

	if be.Backend.Config.Hostname == "" {
		be.Backend.Config.Hostname = be.Cmd.String("host")
	}

	client, err := be.Client()
	if err != nil {
		log.WithError(err).Error("can't get client")
		return nil, err
	}

	limit := be.Cmd.Int("limit")

	pageSize := 100
	if limit > 0 && limit < pageSize {
		pageSize = limit
	}

	// organization, err := be.Organization()
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to get organization: %w", err)
	// }

	workspace, err := be.Workspace()
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace name: %w", err)
	}

	log.Debugf("workspace: %v", workspace)

	options := tfe.RunListForOrganizationOptions{
		WorkspaceNames: workspace.Name,
		ListOptions:    tfe.ListOptions{PageNumber: 1, PageSize: pageSize},
	}

	var results []*tfe.Run

	// Paginate through the dataset
	for {
		page, err := client.Runs.ListForOrganization(be.Ctx, "tfctl", &options)
		if err != nil {
			return nil, err
		}

		results = append(results, page.Items...)

		if len(results) >= limit {
			results = results[:limit]
			break
		}

		log.Debugf("page: %d, total: %d", page.CurrentPage, len(results))

		if page.NextPage == 0 {
			break
		}
		options.ListOptions.PageNumber++
	}

	return results, nil

}

// StateVersions implements backend.Backend.
func (be *BackendRemote) StateVersions() ([]*tfe.StateVersion, error) {
	if len(be.StateVersionList) > 0 {
		log.Errorf("be.StateVersionList: preloaded with %d", len(be.StateVersionList))
		return be.StateVersionList, nil
	}

	if be.Backend.Config.Hostname == "" {
		be.Backend.Config.Hostname = be.Cmd.String("host")
	}

	client, err := be.Client()
	if err != nil {
		log.WithError(err).Error("can't get client")
		return nil, err
	}

	// Short-circuit this if we're in sq but not --diff and no sv is passed. This
	// is the most common sq use case and there's no need to paginate through all
	// the StateVersion records when we know we're always going to take the first
	// one. This makes a noticeable performance difference on slow servers or
	// workspaces with large SV lists.
	diff := be.Cmd.Bool("diff")
	sv := be.Cmd.String("sv")
	limit := be.Cmd.Int("limit")
	if (be.Cmd.Name == "sq" || be.Cmd.Name == "si") && sv == "0" && !diff {
		limit = 1
	}

	pageSize := 100
	if limit > 0 && limit < pageSize {
		pageSize = limit
	}

	organization, err := be.Organization()
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	workspace, err := be.WorkspaceName()
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace name: %w", err)
	}
	options := tfe.StateVersionListOptions{
		Workspace:    workspace,
		Organization: organization,
		ListOptions:  tfe.ListOptions{PageNumber: 1, PageSize: pageSize},
	}

	var results []*tfe.StateVersion

	// Include related resources with each page so callers can access them
	// without issuing per-item reads.
	includes := []tfe.StateVersionIncludeOpt{
		tfe.SVrun,
		tfe.SVoutputs,
	}

	// Paginate through the dataset
	for {
		page, err := listStateVersionsWithInclude(be.Ctx, client, &options, includes)
		if err != nil {
			ctxErr := ErrorContext{
				Host:      be.Backend.Config.Hostname,
				Org:       organization,
				Workspace: workspace,
				Operation: "list state versions",
				Resource:  "stateversion",
			}
			return nil, FriendlyTFE(err, ctxErr)
		}

		results = append(results, page.Items...)

		if len(results) >= limit {
			results = results[:limit]
			break
		}

		log.Debugf("page: %d, total: %d", page.CurrentPage, len(results))

		if page.Pagination.NextPage == 0 {
			break
		}
		options.ListOptions.PageNumber++
	}

	return results, nil
}

// listStateVersionsWithInclude calls the TFE state-versions list endpoint with
// support for the "include" query parameter to load related resources.
func listStateVersionsWithInclude(ctx context.Context, client *tfe.Client, base *tfe.StateVersionListOptions, include []tfe.StateVersionIncludeOpt) (*tfe.StateVersionList, error) {
	// Build a local options struct that mirrors StateVersionListOptions and adds include.
	type listOpts struct {
		tfe.ListOptions
		Organization string                       `url:"filter[organization][name]"`
		Workspace    string                       `url:"filter[workspace][name]"`
		Include      []tfe.StateVersionIncludeOpt `url:"include,omitempty"`
	}
	lo := listOpts{
		ListOptions:  base.ListOptions,
		Organization: base.Organization,
		Workspace:    base.Workspace,
		Include:      include,
	}

	req, err := client.NewRequest("GET", "state-versions", &lo)
	if err != nil {
		return nil, err
	}
	svl := &tfe.StateVersionList{}
	if err := req.Do(ctx, svl); err != nil {
		return nil, err
	}
	return svl, nil
}

func (be *BackendRemote) String() string {
	beCopy := *be
	if beCopy.Backend.Config.Token != nil {
		beCopy.Backend.Config.Token = "********"
	}
	return fmt.Sprintf("ConfigRemote: %+v", beCopy)
}

// Token retrieves the token from the environment variable, config file, or
// the credentials file, in that order.
func (be *BackendRemote) Token() (string, error) {
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

func (be *BackendRemote) Type() (string, error) {
	return be.Backend.Type, nil
}

// Workspace returns the workspace object for the workspace identified in the
// backend.  The workspace object can't be cached because State() is currently
// using it to get the CurrentStateVersion, which may invalidate the cache.
func (be *BackendRemote) Workspace() (*tfe.Workspace, error) {
	wsName, err := be.WorkspaceName()
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace name: %w", err)
	}

	client, err := be.Client()
	if err != nil {
		return nil, fmt.Errorf("failed to get TFE client: %w", err)
	}

	org, err := be.Organization()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve organization: %w", err)
	}
	ctx := context.Background()

	workspace, err := client.Workspaces.Read(ctx, org, wsName)
	if err != nil {
		ctxErr := ErrorContext{
			Host:      be.Backend.Config.Hostname,
			Org:       org,
			Workspace: wsName,
			Operation: "read workspace",
			Resource:  "workspace",
		}
		return nil, FriendlyTFE(err, ctxErr)
	}

	return workspace, nil
}

func (be *BackendRemote) WorkspaceName() (string, error) {
	ws := be.Cmd.String("workspace")
	if ws != "" {
		return ws, nil
	}

	workspaces := be.Backend.Config.Workspaces

	if workspaces.Name != "" && workspaces.Prefix != "" {
		return "", fmt.Errorf("both workspace name and prefix are set: %w", ErrWorkspaceNameAndPrefixBothSet)
	}

	// If it's a "straight name" just return it and go home early.
	if workspaces.Name != "" {
		log.Debugf("workspace name: %s", workspaces.Name)
		return workspaces.Name, nil
	}

	// This is going to be a "prefixed name". If the environment file exists, this
	// is a multi-workspace configuration. The contents of that file along with
	// Prefix are used to determine the actual state file path.
	var env string
	envFile := filepath.Join(be.RootDir, ".terraform/environment")
	if envFileData, err := os.ReadFile(envFile); err == nil {
		env = string(bytes.TrimSpace(envFileData))
	}

	if be.EnvOverride != "" {
		env = be.EnvOverride
	}

	name := workspaces.Prefix + env
	log.Debugf("workspace prefixed name = %s", name)
	return name, nil
}
