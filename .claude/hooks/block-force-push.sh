#!/bin/bash
# Block force pushes to protect shared history.
INPUT=$(cat)
COMMAND=$(echo "$INPUT" | python3 -c "import json,sys; print(json.load(sys.stdin).get('tool_input', {}).get('command', ''))")

# Only block if the command itself is a git push with force flags.
# Anchored to avoid false positives when "git push --force" appears in commit messages.
FIRST_LINE=$(echo "$COMMAND" | head -1 | sed 's/^[[:space:]]*//')
if echo "$FIRST_LINE" | grep -qE '^git push\b' && echo "$COMMAND" | grep -qE '(-f\b|--force\b|--force-with-lease\b)'; then
  echo "Force push is not allowed. Use a regular push, or ask the user to do it explicitly." >&2
  exit 2
fi
