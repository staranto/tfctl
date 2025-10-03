#!/usr/bin/env bash

# Runs Go code quality checks:
# - go vet (with custom printf funcs)
# - shadow analyzer (if installed)
# - staticcheck (if installed)
# - golangci-lint (if installed)
#
# Usage:
#   tools/check.sh [--all] [--hard]
#     --all         Run checks regardless of staged Go file changes.
#     --hard        Enable strict mode (set -euo pipefail). Default is soft mode.
#     (default)     Only run if staged .go files are present, soft mode.

ROOT_DIR=$(git rev-parse --show-toplevel)
cd "$ROOT_DIR"

# Parse flags (order-independent)
MODE="changed-only"
HARD=0
for arg in "$@"; do
  case "$arg" in
    --all)
      MODE="all"
      ;;
    --hard)
      HARD=1
      ;;
  esac
done

# Enable strict mode only if requested
if [[ "$HARD" -eq 1 ]]; then
  set -euo pipefail
fi

go_changed() {
  git diff --cached --name-only --diff-filter=ACMR | grep -E '\.go$' || true
}

if [[ "$MODE" == "changed-only" ]]; then
  CHANGED_FILES=$(go_changed)
  if [[ -z "$CHANGED_FILES" ]]; then
    echo "[pre-commit] no staged .go changes; skipping Go quality checks." >&2
    exit 0
  fi
  echo "[pre-commit] staged .go changes detected:" >&2
  echo "$CHANGED_FILES" | sed 's/^/  - /' >&2
fi

echo "[pre-commit] running Go code quality checks..." >&2

# 1) go vet with custom printf funcs so apex/log *f calls are validated
echo "[pre-commit] running Go Vet..." >&2
go vet -printfuncs='Debugf,Infof,Warnf,Errorf' ./...

# 2) shadow analyzer (optional)
echo "[pre-commit] running shadow..." >&2
if command -v shadow >/dev/null 2>&1; then
  shadow ./...
else
  echo "[pre-commit] shadow not found; skipping. Install with: go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest" >&2
fi

# 3) staticcheck (optional)
echo "[pre-commit] running staticcheck..." >&2
if command -v staticcheck >/dev/null 2>&1; then
  staticcheck ./...
else
  echo "[pre-commit] staticcheck not found; skipping. Install with: go install honnef.co/go/tools/cmd/staticcheck@latest" >&2
fi

# 4) golangci-lint (optional)
echo "[pre-commit] running golangci-lint..." >&2
if command -v golangci-lint >/dev/null 2>&1; then
  golangci-lint run
else
  echo "[pre-commit] golangci-lint not found; skipping. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" >&2
fi

echo "[pre-commit] Go code quality checks passed." >&2
