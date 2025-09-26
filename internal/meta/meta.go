// Copyright © 2025 Steve Taranto staranto@gmail.com
// SPDX-License-Identifier: MIT

package meta

import (
	"context"

	"github.com/staranto/tfctlgo/internal/config"
)

type RootDirSpec struct {
	RootDir string
	Env     string
}

// Meta are the meta-options that are available on all or most commands.
type Meta struct {
	Args    []string
	Config  config.ConfigType
	Context context.Context
	RootDirSpec
	StartingDir string
}
