#!/usr/bin/env bash
#MISE description="Appends a new tag"
set -euo pipefail

today=$(date +"%Y.%m.%d")

# Get today's tags
tags=$(git tag --list "${today}.*")

if [ -z "$tags" ]; then
    next=1
else
    last=$(echo "$tags" | sed "s/${today}\.//" | sort -n | tail -1)
    next=$((last + 1))
fi

new_tag="${today}.${next}"

echo "Appending tag: $new_tag"
git tag "$new_tag"
