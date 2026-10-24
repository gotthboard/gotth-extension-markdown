#!/bin/sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd -P)
. "$root/scripts/pinned-toolchain.sh"
[ "$#" -eq 0 ] || { echo "usage: $0" >&2; exit 2; }
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT HUP INT TERM
cd "$root"
sha256sum -c scripts/sanitized-run.sha256
CGO_ENABLED=0 go build -mod=readonly -trimpath -buildvcs=false -ldflags='-s -w -buildid=' -o "$tmp/sanitized-run" ./scripts/sanitized-run.go
cmp scripts/sanitized-run "$tmp/sanitized-run"
/usr/bin/ldd scripts/sanitized-run > "$tmp/ldd" 2>&1 || true
grep -Fq 'not a dynamic executable' "$tmp/ldd"
find /tmp -maxdepth 1 -type d -name 'gm.*' -printf '%f\n' | LC_ALL=C sort > "$tmp/before"
set +e
scripts/sanitized-run /bin/sh -c 'exit 23'
status=$?
set -e
[ "$status" -eq 23 ] || { echo "sanitized runner did not preserve failure status" >&2; exit 1; }
find /tmp -maxdepth 1 -type d -name 'gm.*' -printf '%f\n' | LC_ALL=C sort > "$tmp/after"
cmp "$tmp/before" "$tmp/after" || { echo "failed sanitized run leaked its temporary root" >&2; exit 1; }
printf '%s\n' 'static sanitized runner reproduced byte-for-byte and failure cleanup verified'
