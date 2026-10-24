#!/bin/sh
set -eu
PATH=/usr/lib/go/bin:/usr/bin:/bin; export PATH
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd -P)
. "$root/scripts/pinned-toolchain.sh"
tmp=$(mktemp -d)
canary="$root/.source-manifest-hostile-canary"
option_canary="$root/--version"
trap 'rm -rf "$tmp"; rm -f "$canary" "$option_canary" "$root/.source-manifest-in-tree-test"' EXIT HUP INT TERM
manifest="$tmp/source-manifest.sha256"
if (cd "$root" && "$root/scripts/source-manifest.sh" .source-manifest-in-tree-test >/dev/null 2>&1); then
  echo 'source manifest accepted an in-tree relative output' >&2
  exit 1
fi
[ ! -e "$root/.source-manifest-in-tree-test" ]
"$root/scripts/source-manifest.sh" "$manifest"
printf '%s\n' drift > "$canary"
if "$root/scripts/verify-source-snapshot.sh" "$manifest" >/dev/null 2>&1; then
  echo 'source verifier accepted an added file' >&2
  exit 1
fi
rm -f "$canary"
printf '%s\n' option-safe > "$option_canary"
"$root/scripts/source-manifest.sh" "$manifest"
grep -Eq '^[0-9a-f]{64}  --version$' "$manifest" || { echo 'source manifest mishandled an option-like filename' >&2; exit 1; }
"$root/scripts/verify-source-snapshot.sh" "$manifest"
rm -f "$option_canary"
"$root/scripts/source-manifest.sh" "$manifest"
"$root/scripts/verify-source-snapshot.sh" "$manifest"
printf '%s\n' 'exact source fileset and option-like filename handling verified'
