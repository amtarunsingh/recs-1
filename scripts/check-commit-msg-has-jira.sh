#!/usr/bin/env bash
set -euo pipefail

COMMIT_MSG_FILE="$1"

if ! grep -qE ' \[[A-Z]+-\d+\]$' "$COMMIT_MSG_FILE"; then
  echo "❌ Conventional Commit message must end with a Jira ticket number"
  echo "e.g. feat(votes): add vote type validation for crush votes [RECS-1234]"
  exit 1
fi