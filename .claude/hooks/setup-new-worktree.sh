#!/bin/bash
# Copies gitignored sprite assets into a new worktree so go:embed targets compile.
# go:embed does not follow symlinks, so assets must be physically present in the worktree.
set -euo pipefail

INPUT=$(cat)
WORKTREE=$(echo "$INPUT" | python3 -c "import json,sys; print(json.load(sys.stdin)['worktree_path'])")

if [ -z "$WORKTREE" ]; then
  echo "setup-new-worktree: could not read worktree_path from hook input" >&2
  exit 1
fi

SPRITES="${CLAUDE_PROJECT_DIR}/assets/sprites"
if [ -d "$SPRITES" ]; then
  mkdir -p "$WORKTREE/assets/sprites"
  cp -r "$SPRITES/." "$WORKTREE/assets/sprites/"
  echo "Copied sprites to $WORKTREE/assets/sprites/"
fi
