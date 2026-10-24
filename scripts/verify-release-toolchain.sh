#!/bin/sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd -P)
. "$root/scripts/pinned-toolchain.sh"
[ "$#" -eq 0 ] || { echo "usage: $0" >&2; exit 2; }
expected="$root/scripts/release-toolchain.sha256"
dynamic_expected="$root/scripts/release-toolchain-dynamic.paths"
[ -f "$expected" ] || { echo "release toolchain manifest unavailable" >&2; exit 1; }
[ -f "$dynamic_expected" ] || { echo "release dynamic-closure manifest unavailable" >&2; exit 1; }
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT HUP INT TERM
(cd "$root" && sha256sum -c scripts/sanitized-run.sha256)
/usr/bin/ldd "$root/scripts/sanitized-run" > "$tmp/runner.ldd" 2>&1 || true
grep -Fq 'not a dynamic executable' "$tmp/runner.ldd" || { echo "sanitized runner is not static" >&2; exit 1; }
"$root/scripts/release-toolchain-dynamic-paths.sh" > "$tmp/dynamic.actual"
cmp "$dynamic_expected" "$tmp/dynamic.actual" || { echo "release dynamic-library resolution drifted" >&2; exit 1; }
{
  "$root/scripts/release-toolchain-paths.sh"
  while IFS= read -r dynamic_path; do
    [ -n "$dynamic_path" ] || continue
    printf '%s\0' "$dynamic_path"
  done < "$dynamic_expected"
} | LC_ALL=C sort -zu > "$tmp/paths"
xargs -0 sha256sum < "$tmp/paths" > "$tmp/actual.unsorted"
LC_ALL=C sort -k2 "$tmp/actual.unsorted" > "$tmp/actual"
cmp "$expected" "$tmp/actual" || { echo "release toolchain fileset or content drifted" >&2; exit 1; }
manifest_sha=$(sha256sum "$expected"); manifest_sha=${manifest_sha%% *}
printf 'release toolchain closure verified: files=%s manifest_sha256=%s fresh_gocache=%s\n' "$(wc -l < "$expected")" "$manifest_sha" "$GOCACHE"
