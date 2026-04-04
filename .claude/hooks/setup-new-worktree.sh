#!/bin/bash
# Creates a git worktree and copies gitignored sprite assets into it so
# go:embed targets compile. go:embed does not follow symlinks, so assets
# must be physically present in the worktree.
set -euo pipefail

INPUT=$(cat)
# The WorktreeCreate payload has no path field; derive the path from the worktree name.
# Worktrees are always created at $CLAUDE_PROJECT_DIR/.claude/worktrees/<name>.
NAME=$(echo "$INPUT" | python3 -c "import json,sys; print(json.load(sys.stdin)['name'])")
WORKTREE="${CLAUDE_PROJECT_DIR}/.claude/worktrees/${NAME}"

if [ -z "$NAME" ]; then
  echo "setup-new-worktree: could not read name from hook input" >&2
  exit 1
fi

# Create the git worktree on a new branch.
mkdir -p "${CLAUDE_PROJECT_DIR}/.claude/worktrees"
git -C "$CLAUDE_PROJECT_DIR" worktree add "$WORKTREE" -b "$NAME" HEAD >&2

# Copy gitignored sprite assets so go:embed targets compile.
SPRITES="${CLAUDE_PROJECT_DIR}/assets/sprites"
if [ -d "$SPRITES" ]; then
  mkdir -p "$WORKTREE/assets/sprites"
  cp -r "$SPRITES/." "$WORKTREE/assets/sprites/"
  echo "Copied sprites to $WORKTREE/assets/sprites/" >&2
fi

# Output the worktree path on stdout — the tool uses this as the handoff signal.
echo "$WORKTREE"
