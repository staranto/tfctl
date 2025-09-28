package command

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/staranto/tfctlgo/internal/meta"
	"github.com/urfave/cli/v3"
)

const bashCompletionScript = `# bash completion for tfctl
# Fallback if bash-completion is not installed
if ! declare -F _get_comp_words_by_ref >/dev/null 2>&1; then
  _get_comp_words_by_ref() {
    cur=${COMP_WORDS[COMP_CWORD]}
    prev=${COMP_WORDS[COMP_CWORD-1]}
  }
fi

_tfctl()
{
    local cur prev cmd
    COMPREPLY=()
    _get_comp_words_by_ref -n : cur prev

    if [[ ${COMP_CWORD} -eq 1 ]]; then
        COMPREPLY=( $(compgen -W "mq oq pq rq si sq svq wq completion --help --version" -- "$cur") )
        return 0
    fi

    cmd=${COMP_WORDS[1]}
  local common="--attrs -a --color -c --filter -f --output -o --sort -s --titles -t --tldr"

    # Determine if an optional RootDir (first non-flag after subcommand) has already been provided
    local have_rootdir=0
    local idx=2
    while [[ $idx -lt ${#COMP_WORDS[@]} ]]; do
        local w=${COMP_WORDS[$idx]}
        if [[ $w != -* ]]; then
            have_rootdir=1
            break
        fi
        ((idx++))
    done

    case "$cmd" in
    mq)
      local opts="$common --schema --host -h --org"
            ;;
        oq)
      local opts="$common --schema --host -h"
            ;;
        pq)
      local opts="$common --schema --host -h --org"
            ;;
        rq)
      local opts="$common --schema --host -h --org --limit -l --workspace -w"
            ;;
        si)
            local opts="--passphrase -p --sv"
            ;;
        sq)
      local opts="$common --host -h --org --workspace -w --concrete -k --diff --noshort --passphrase --sv --limit"
            ;;
        svq)
      local opts="$common --schema --host -h --org --limit -l --workspace -w"
            ;;
        wq)
      local opts="$common --schema --host -h --org"
            ;;
        completion)
            local opts="bash zsh"
            COMPREPLY=( $(compgen -W "$opts" -- "$cur") )
            return 0
            ;;
        *)
            local opts="$common"
            ;;
    esac

    if [[ "$prev" == "--output" || "$prev" == "-o" ]]; then
        COMPREPLY=( $(compgen -W "text json raw yaml" -- "$cur") )
        return 0
    fi

  # If current token starts with '-', or we've already consumed RootDir, offer flags
  if [[ "$cur" == -* || $have_rootdir -eq 1 ]]; then
    COMPREPLY=( $(compgen -W "$opts" -- "$cur") )
    return 0
  fi

  # Otherwise, we're on the (optional) RootDir positional — complete directories
  COMPREPLY=( $(compgen -o dirnames -- "$cur") )
  return 0
}

complete -F _tfctl tfctl
`

const zshCompletionScript = `#compdef tfctl

_tfctl() {
  local -a cmds
  cmds=(
    'mq:module registry query'
    'oq:organization query'
    'pq:project query'
    'rq:run query'
    'si:interactive state inspector'
    'sq:state query'
    'svq:state version query'
    'wq:workspace query'
    'completion:generate shell completion script'
  )

  local -a common
  common=(
  '(-a --attrs)'{-a,--attrs}'[attributes to include]:attrs'
  '(-c --color)'{-c,--color}'[enable colored text]'
  '(-f --filter)'{-f,--filter}'[filters to apply]:filters'
  '(-o --output)'{-o,--output}'[output format]:format:(text json raw yaml)'
  '(-s --sort)'{-s,--sort}'[sort attributes]:attrs'
  '(-t --titles)'{-t,--titles}'[show titles]'
  '--tldr[show tldr page]'
  )

  if (( CURRENT == 2 )); then
    _describe -t commands 'tfctl commands' cmds
    return
  fi

  local curcontext="$curcontext" state line
  case $words[2] in
    mq)
      _arguments -C \
        $common \
        '--schema[dump schema]' \
        '(-h --host)'{-h,--host}'[host]' \
        '--org[organization]' \
        '::RootDir:_directories'
      ;;
    oq)
      _arguments -C \
        $common \
        '--schema[dump schema]' \
        '(-h --host)'{-h,--host}'[host]' \
        '::RootDir:_directories'
      ;;
    pq)
      _arguments -C \
        $common \
        '--schema[dump schema]' \
        '(-h --host)'{-h,--host}'[host]' \
        '--org[organization]' \
        '::RootDir:_directories'
      ;;
    rq)
      _arguments -C \
        $common \
        '--schema[dump schema]' \
        '--limit[-l][limit results]':limit \
        '(-h --host)'{-h,--host}'[host]' \
        '--org[organization]' \
        '::RootDir:_directories'
      ;;
    si)
      _arguments -C \
        '(-p --passphrase)'{-p,--passphrase}'[state passphrase]' \
        '--sv[state version]' \
        '::RootDir:_directories'
      ;;
    sq)
      _arguments -C \
        $common \
        '--concrete[only concrete resources]' \
        '--diff[diff between state versions]' \
        '--noshort[full resource paths]' \
        '--passphrase[state passphrase]' \
        '--sv[state version]' \
        '--limit[limit results]' \
        '(-h --host)'{-h,--host}'[host]' \
        '--org[organization]' \
        '(-w --workspace)'{-w,--workspace}'[workspace]' \
        '::RootDir:_directories'
      ;;
    svq)
      _arguments -C \
        $common \
        '--schema[dump schema]' \
        '--limit[-l][limit results]' \
        '(-h --host)'{-h,--host}'[host]' \
        '--org[organization]' \
        '(-w --workspace)'{-w,--workspace}'[workspace]' \
        '::RootDir:_directories'
      ;;
    wq)
      _arguments -C \
        $common \
        '--schema[dump schema]' \
        '(-h --host)'{-h,--host}'[host]' \
        '--org[organization]' \
        '*:directory:_directories'
      ;;
    completion)
      _arguments '1: :((bash zsh))'
      ;;
    *)
      _arguments -C $common '*:directory:_directories'
      ;;
  esac
}

# If this file is sourced directly (not autoloaded via fpath), ensure compsys is initialized and register the completion
if ! typeset -f compdef >/dev/null 2>&1; then
  autoload -Uz compinit && compinit -i
fi
compdef _tfctl tfctl tfctlgo
`

func CompletionCommandAction(ctx context.Context, cmd *cli.Command) error {
	shell := ""
	if args := cmd.Args().Slice(); len(args) > 0 {
		shell = args[0]
	}
	switch shell {
	case "bash":
		fmt.Fprint(os.Stdout, bashCompletionScript)
	case "zsh":
		fmt.Fprint(os.Stdout, zshCompletionScript)
	default:
		// Try to detect from SHELL or print help
		sh := os.Getenv("SHELL")
		if strings.HasSuffix(sh, "zsh") {
			fmt.Fprint(os.Stdout, zshCompletionScript)
		} else if strings.HasSuffix(sh, "bash") {
			fmt.Fprint(os.Stdout, bashCompletionScript)
		} else {
			fmt.Fprintln(os.Stderr, "usage: tfctl completion [bash|zsh]")
			return nil
		}
	}
	return nil
}

func CompletionCommandBuilder(cmd *cli.Command, meta meta.Meta) *cli.Command {
	return &cli.Command{
		Name:      "completion",
		Usage:     "generate shell completion script",
		UsageText: "tfctl completion [bash|zsh]",
		Metadata: map[string]any{
			"meta": meta,
		},
		Action: CompletionCommandAction,
	}
}
