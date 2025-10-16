#!/bin/bash

if [ $# -lt 1 ] || [ $# -gt 2 ]; then
  echo "Usage: $0 <days>"
  exit 1
fi

DAYS="$1"

# If a repo is provided as $2, use it; otherwise derive from git remote.
if [ $# -ge 2 ]; then
  REPO="$2"
else
  REPO=$(git remote get-url origin | sed -n 's#.*github.com[:/]\([^/]\+\)/\([^/.]\+\).*#\1/\2#p')
fi
if [ -z "$REPO" ]; then
  echo "Could not determine GitHub repo. Provide as [owner/repo] or set a valid git remote."
  exit 2
fi

SECS_AGO=$(date -d "$DAYS days ago" +%s)

echo "== Deleting releases older than $DAYS days ($SECS_AGO) in $REPO =="

gh release list --repo "$REPO" --limit 1000 | while read -r line; do
  # Skip entries marked as Latest (case-insensitive)
  if echo "$line" | grep -qi 'latest'; then
    continue
  fi
  tag=$(echo "$line" | awk '{print $1}')
  date=$(echo "$line" | awk '{print $3}')
  release_epoch=$(date -d "$date" +%s)
  if [ "$release_epoch" -lt "$SECS_AGO" ]; then
    echo "Deleting release: $tag ($date)"
    gh release delete "$tag" --repo "$REPO" --yes
  fi
done

echo "== Deleting tags older than $DAYS days in $REPO (local & remote) =="

git fetch --tags
git for-each-ref --sort=creatordate refs/tags --format '%(refname:short) %(creatordate:iso8601)' | while read -r tag tdate; do
  tag_epoch=$(date -d "$tdate" +%s)
  if [ "$tag_epoch" -lt "$SECS_AGO" ]; then
    echo "Deleting tag: $tag ($tdate)"
    git tag -d "$tag"
    git push --delete origin "$tag"
  fi
done