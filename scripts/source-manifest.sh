#!/bin/sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd -P)
. "$root/scripts/pinned-toolchain.sh"
if [ "$#" -ne 1 ]; then echo "usage: $0 <external-output-file>" >&2; exit 2; fi
output_input=$1
output_dir=$(CDPATH= cd -- "$(dirname -- "$output_input")" && pwd -P)
output="$output_dir/$(basename -- "$output_input")"
case $output in "$root"|"$root"/*) echo "source manifest output must be outside the source tree" >&2; exit 2;; esac
tmp=$(mktemp -d)
output_tmp=$(mktemp "$output_dir/.source-manifest.XXXXXX")
trap 'rm -rf "$tmp"; rm -f "$output_tmp"' EXIT HUP INT TERM
cd "$root"
if find . -path './.git' -prune -o -type l -print -quit | grep -q .; then
  echo "source snapshot must not contain symbolic links" >&2
  exit 1
fi
find . -path './.git' -prune -o -type f -printf '%P\0' > "$tmp/paths.unsorted"
LC_ALL=C sort -z < "$tmp/paths.unsorted" > "$tmp/paths"
[ -s "$tmp/paths" ] || { echo "candidate fileset is empty" >&2; exit 1; }
xargs -0 sha256sum -- < "$tmp/paths" > "$tmp/manifest"
LC_ALL=C sort -k2 < "$tmp/manifest" > "$output_tmp"
mv -f "$output_tmp" "$output"
