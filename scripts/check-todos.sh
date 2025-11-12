#!/usr/bin/env bash
set -euo pipefail

VALID_TODO_REGEX="(TODO\[[A-Z]+-\d+\]|\.TODO\(\))"

echo "🔍 checking TODOs…"

# 1) get all TODO lines (tracked files only)
all_todos=$(git grep -n -I -E 'TODO' -- '*.go' || true)

# nothing to check
[ -z "$all_todos" ] && { echo "✅ no TODOs found"; exit 0; }

# 2) exclude TODOs that already have a ticket
bad_todos=$(printf '%s\n' "$all_todos" | grep -Ev "$VALID_TODO_REGEX" || true)

if [ -n "$bad_todos" ]; then
  echo "❌ TODOs missing ticket numbers:"
  echo "$bad_todos"
  echo
  echo "Ticket number should be specified like: TODO[KEY-1234]"
  exit 1
fi

echo "✅ all TODOs include ticket numbers"