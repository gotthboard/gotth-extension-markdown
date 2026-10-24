#!/bin/sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd -P)
. "$root/scripts/pinned-toolchain.sh"
[ "$#" -eq 0 ] || { echo "usage: $0" >&2; exit 2; }
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT HUP INT TERM
cd "$root"
if find . -path './.git' -prune -o -name .gitattributes -print -quit | grep -q .; then
  echo "candidate diff source forbids .gitattributes" >&2
  exit 1
fi
git bundle create "$tmp/source.bundle" HEAD
git clone --no-checkout --no-local "$tmp/source.bundle" "$tmp/repo" >/dev/null 2>&1
find . -path './.git' -prune -o -type f -printf '%P\0' | LC_ALL=C sort -z > "$tmp/paths"
tar --null --verbatim-files-from -T "$tmp/paths" -cf "$tmp/candidate.tar"
tar --same-permissions -xf "$tmp/candidate.tar" -C "$tmp/repo"
git -C "$tmp/repo" -c core.attributesFile=/dev/null -c core.excludesFile=/dev/null -c core.fsmonitor=false add -A -- .
git -C "$tmp/repo" -c core.attributesFile=/dev/null -c core.excludesFile=/dev/null -c core.fsmonitor=false diff --cached --check HEAD -- . ':(exclude)third_party/licenses/**'
printf '%s\n' 'candidate-wide diff check passed; byte-exact upstream legal files are explicitly exempt and independently verified'
