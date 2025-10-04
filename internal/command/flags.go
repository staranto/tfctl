// Copyright © 2025 Steve Taranto staranto@gmail.com
// SPDX-License-Identifier: MIT

package command

import (
	"os/exec"

	altsrc "github.com/urfave/cli-altsrc/v3"
	yaml "github.com/urfave/cli-altsrc/v3/yaml"
	"github.com/urfave/cli/v3"

	"github.com/staranto/tfctlgo/internal/config"
)

func init() {
	cfg, _ = config.Load("")
}

var (
	cfg      config.Type
	tldrFlag *cli.BoolFlag = &cli.BoolFlag{
		Name:        "tldr",
		Usage:       "show tldr page",
		Hidden:      !pathHas("tldr"),
		HideDefault: true,
	}

	schemaFlag *cli.BoolFlag = &cli.BoolFlag{
		Name:        "schema",
		Usage:       "dump the schema",
		HideDefault: true,
	}

	workspaceFlag *cli.StringFlag = &cli.StringFlag{
		Name:    "workspace",
		Aliases: []string{"w"},
		Usage:   "workspace to use for query. Overrides the backend",
		Sources: cli.NewValueSourceChain(
			cli.EnvVar("TFCTL_WORKSPACE"),
		),
		Value: "",
	}
)

func NewGlobalFlags(params ...string) (flags []cli.Flag) {
	flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "attrs",
			Aliases: []string{"a"},
			Usage:   "comma-separated list of attributes to include in results",
		},
		&cli.BoolFlag{
			Name:        "color",
			Aliases:     []string{"c"},
			Usage:       "enable colored text output",
			HideDefault: true,
			Sources: cli.NewValueSourceChain(
				yaml.YAML(params[0]+"."+"color", altsrc.StringSourcer(cfg.Source)),
				yaml.YAML("color", altsrc.StringSourcer(cfg.Source)),
			),
		},
		&cli.StringFlag{
			Name:    "filter",
			Aliases: []string{"f"},
			Usage:   "comma-separated list of filters to apply to results",
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "output format",
			Sources: cli.NewValueSourceChain(
				yaml.YAML(params[0]+"."+"output", altsrc.StringSourcer(cfg.Source)),
				yaml.YAML("output", altsrc.StringSourcer(cfg.Source)),
			),
			Value: "text",
			Validator: func(value string) error {
				return FlagValidators(value, OutputValidator)
			},
		},
		&cli.StringFlag{
			Name:    "sort",
			Aliases: []string{"s"},
			Usage:   "comma-separated list of attributes to sort the results by",
			Sources: cli.NewValueSourceChain(
				yaml.YAML(params[0]+"."+"sort", altsrc.StringSourcer(cfg.Source)),
			),
		},
		&cli.BoolFlag{
			Name:        "titles",
			Aliases:     []string{"t"},
			Usage:       "show titles with text output",
			HideDefault: true,
			Sources: cli.NewValueSourceChain(
				yaml.YAML(params[0]+"."+"titles", altsrc.StringSourcer(cfg.Source)),
				yaml.YAML("titles", altsrc.StringSourcer(cfg.Source)),
			),
		},
	}

	return
}

func NewHostFlag(params ...string) (flag *cli.StringFlag) {
	flag = &cli.StringFlag{
		Name:    "host",
		Aliases: []string{"h"},
		Usage:   "host to use for all commands. Overrides the backend",
		Sources: cli.NewValueSourceChain(
			cli.EnvVar("TFCTL_HOST"),
			cli.EnvVar("TF_CLOUD_HOSTNAME"),
		),
		Value: "app.terraform.io",
	}

	if len(params) == 2 {
		flag = NameSpacedValueChainFlag(params[0], params[1], flag)
	}

	return
}

func NewOrgFlag(params ...string) (flag *cli.StringFlag) {
	flag = &cli.StringFlag{
		Name:  "org",
		Usage: "organization to use for all commands. Overrides the backend",
		Sources: cli.NewValueSourceChain(
			cli.EnvVar("TFCTL_ORG"),
			cli.EnvVar("TF_CLOUD_ORGANIZATION"),
		),
	}

	if len(params) == 2 {
		flag = NameSpacedValueChainFlag(params[0], params[1], flag)
	}

	return
}

func NameSpacedValueChainFlag(ns string, path string, flag *cli.StringFlag) *cli.StringFlag {
	src := yaml.YAML(ns+"."+flag.Name, altsrc.StringSourcer(path))
	flag.Sources.Chain = append(flag.Sources.Chain, src)

	src = yaml.YAML(flag.Name, altsrc.StringSourcer(path))
	flag.Sources.Chain = append(flag.Sources.Chain, src)

	return flag
}

// pathHas checks if the given key exists in cfg.Source.
func pathHas(target string) bool {
	_, err := exec.LookPath(target)
	return err == nil
}
