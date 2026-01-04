#!/bin/bash
# write-file.sh - Write content to a file with path validation
#
# Usage: write-file.sh <path>
# Content is read from stdin
#
# Guardrails:
# - Path must be absolute
# - Path must be within allowed directories (.claude/ui, .claude)
# - Creates parent directories automatically

set -e

FILE_PATH="$1"

# Validate path is provided
if [ -z "$FILE_PATH" ]; then
    echo "Error: No file path provided" >&2
    echo "Usage: write-file.sh <absolute-path>" >&2
    exit 1
fi

# Validate path is absolute
if [[ "$FILE_PATH" != /* ]]; then
    echo "Error: Path must be absolute (start with /)" >&2
    exit 1
fi

# Resolve to canonical path (handles ..)
RESOLVED_PATH=$(realpath -m "$FILE_PATH")

# Get the working directory
WORK_DIR=$(pwd)

# Allowed directory patterns (relative to working directory)
ALLOWED_PATTERNS=(
    "$WORK_DIR/.claude/ui/"
    "$WORK_DIR/.claude/"
    "/tmp/"
)

# Check if path is within allowed directories
ALLOWED=false
for pattern in "${ALLOWED_PATTERNS[@]}"; do
    if [[ "$RESOLVED_PATH" == "$pattern"* ]]; then
        ALLOWED=true
        break
    fi
done

if [ "$ALLOWED" = false ]; then
    echo "Error: Path not in allowed directories" >&2
    echo "Allowed: .claude/ui/, .claude/, /tmp/" >&2
    echo "Got: $RESOLVED_PATH" >&2
    exit 1
fi

# Create parent directory
mkdir -p "$(dirname "$RESOLVED_PATH")"

# Write content from stdin
cat > "$RESOLVED_PATH"

echo "Written: $RESOLVED_PATH"
