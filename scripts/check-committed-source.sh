#!/bin/sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd -P)
. "$root/scripts/pinned-toolchain.sh"
[ "$#" -eq 0 ] || { echo "usage: $0" >&2; exit 2; }
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT HUP INT TERM
cd "$root"
find . -path './.git' -prune -o -type f -printf '%P\0' | LC_ALL=C sort -z > "$tmp/files"
git -c core.attributesFile=/dev/null -c core.excludesFile=/dev/null -c core.fsmonitor=false ls-files -z | LC_ALL=C sort -z > "$tmp/tracked"
cmp "$tmp/tracked" "$tmp/files" || { echo "stable release source must contain exactly the committed files" >&2; exit 1; }
if git -c core.attributesFile=/dev/null -c core.excludesFile=/dev/null -c core.fsmonitor=false ls-files -v | grep -Eq '^[a-zS] '; then
  echo "stable release source forbids assume-unchanged and skip-worktree index flags" >&2
  exit 1
fi
git -c core.attributesFile=/dev/null -c core.excludesFile=/dev/null -c core.fsmonitor=false diff --no-ext-diff --quiet HEAD --
git -c core.attributesFile=/dev/null -c core.excludesFile=/dev/null -c core.fsmonitor=false diff --no-ext-diff --cached --quiet HEAD --
printf '%s\n' 'committed-source eligibility verified; snapshot bytes are independently manifest-bound'
