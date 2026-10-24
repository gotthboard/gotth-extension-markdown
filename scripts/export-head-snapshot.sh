#!/bin/sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd -P)
. "$root/scripts/pinned-toolchain.sh"
[ "$#" -eq 2 ] || { echo "usage: $0 <head-commit> <empty-destination-directory>" >&2; exit 2; }
head_commit=$1
destination=$2
case $head_commit in ''|*[!0-9a-f]*) echo "canonical commit ID required" >&2; exit 2;; esac
[ "$(printf %s "$head_commit" | wc -c)" -eq 40 ] || { echo "full SHA-1 commit ID required" >&2; exit 2; }
[ -d "$destination" ] && [ -z "$(find "$destination" -mindepth 1 -print -quit)" ] || { echo "empty destination directory required" >&2; exit 2; }
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT HUP INT TERM
git -C "$root" bundle create "$tmp/source.bundle" HEAD
git clone --bare --no-local "$tmp/source.bundle" "$tmp/source.git" >/dev/null 2>&1
[ "$(git --git-dir="$tmp/source.git" rev-parse HEAD)" = "$head_commit" ] || { echo "bundled HEAD identity changed" >&2; exit 1; }
if git --git-dir="$tmp/source.git" ls-tree -r --name-only "$head_commit" | grep -Eq '(^|/)\.gitattributes$'; then
  echo "stable release source forbids tracked .gitattributes" >&2
  exit 1
fi
git -c core.attributesFile=/dev/null -c core.excludesFile=/dev/null --git-dir="$tmp/source.git" archive --format=tar "$head_commit" > "$tmp/source.tar"
tar --same-permissions -xf "$tmp/source.tar" -C "$destination"
[ "$(git -C "$root" rev-parse HEAD)" = "$head_commit" ] || { echo "source HEAD changed during immutable export" >&2; exit 1; }
printf 'immutable isolated HEAD export verified: %s\n' "$head_commit"
