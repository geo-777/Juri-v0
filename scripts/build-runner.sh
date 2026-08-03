#!/bin/bash

set -e

# Get project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

RUNNER_DIR="$PROJECT_DIR/scripts/runner_script"
RUNNER_BIN="$RUNNER_DIR/runner"

echo "Building runner..."

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
go build -o "$RUNNER_BIN" "$RUNNER_DIR"

echo "Copying runner..."

for lang in c cpp java python; do
    cp "$RUNNER_BIN" "$PROJECT_DIR/docker/$lang/runner"
    chmod +x "$PROJECT_DIR/docker/$lang/runner"
    echo "✓ $lang"
done

echo ""
echo "Runner copied to all Docker contexts."