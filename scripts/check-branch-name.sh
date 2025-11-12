#!/usr/bin/env bash
set -euo pipefail

BRANCH_REGEX="^[A-Z]+-[0-9]+(-[a-z0-9]+)*$"

branch_name=$(git rev-parse --abbrev-ref HEAD)

echo "🔍 Checking branch name: $branch_name"

if ! echo "$branch_name" | grep -qE "$BRANCH_REGEX" ; then
  echo "❌ Invalid branch name!"
  echo "Expected pattern: JIRA-TICKET-NUMBER-description"
  echo "Example: RECS-1234"
  echo "Example: RECS-1234-add-user-endpoint"
  exit 1
fi

echo "✅ Branch name OK: $branch_name"