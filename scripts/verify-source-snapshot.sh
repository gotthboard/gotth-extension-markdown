#!/bin/sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd -P)
. "$root/scripts/pinned-toolchain.sh"
if [ "$#" -ne 1 ]; then echo "usage: $0 <expected-source-manifest>" >&2; exit 2; fi
expected=$(CDPATH= cd -- "$(dirname -- "$1")" && pwd -P)/$(basename -- "$1")
[ -f "$expected" ] || { echo "expected source manifest is unavailable" >&2; exit 1; }
actual=$(mktemp)
trap 'rm -f "$actual"' EXIT HUP INT TERM
"$root/scripts/source-manifest.sh" "$actual"
if ! cmp "$expected" "$actual"; then
  echo "source fileset/content does not match the bound manifest" >&2
  diff -u "$expected" "$actual" >&2 || true
  exit 1
fi
