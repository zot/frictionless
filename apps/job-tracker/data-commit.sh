#!/bin/bash
# Commit changes to data.fossil
# Usage: ./data-commit.sh [message]

STORAGE_DIR="$(realpath -s "$(dirname "$0")/../../storage/job-tracker/data")"
FOSSIL_BIN="$HOME/.claude/bin/fossil"
MESSAGE="${1:-Update data}"

cd "$STORAGE_DIR"
"$FOSSIL_BIN" addremove --ignore "*" 2>/dev/null || true
"$FOSSIL_BIN" add data.json jobs . 2>/dev/null || true
"$FOSSIL_BIN" commit -m "$MESSAGE" --no-warnings 2>/dev/null || true
