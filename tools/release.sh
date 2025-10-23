#!/bin/bash

if [ $# -lt 1 ] || [ $# -gt 2 ]; then
  echo "Usage: $0 <semver>"
  exit 1
fi

SEMVER="$1"

git push origin --delete "${SEMVER}"
git tag --delete "${SEMVER}"
git tag "${SEMVER}"
git push origin
git push origin "${SEMVER}"