#!/bin/bash
# Verification gate: block agent stop if there are uncommitted changes and checks fail.
# Only runs when there is actual work to verify — skips clean working trees.
# Excludes e2e_tests: they require a display and panic in subprocess/headless contexts.

if git diff --quiet && git diff --cached --quiet; then
  exit 0
fi

PKGS=$(go list ./... 2>/dev/null | grep -v '/e2e_tests')

if ! make lint 2>&1; then
  echo "lint failed — fix before stopping" >&2
  exit 2
fi

if ! go test -race $PKGS 2>&1; then
  echo "tests failed — fix before stopping" >&2
  exit 2
fi
