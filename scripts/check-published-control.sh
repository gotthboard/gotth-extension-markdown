#!/bin/sh
set -eu
PATH=/usr/lib/go/bin:/usr/bin:/bin; export PATH
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd -P)
. "$root/scripts/pinned-toolchain.sh"
. "$root/scripts/rg-policy-lib.sh"
cd "$root"
grep -Eq '^[[:space:]]+github\.com/gotthboard/gotth-extensions v1\.0\.0$' go.mod
grep -Eq '^[[:space:]]+github\.com/gotthboard/gotth-sdk v1\.0\.0$' go.mod
! grep -Eq '^[[:space:]]*replace[[:space:]]|^replace[[:space:]]*\(' go.mod
grep -Fqx 'github.com/gotthboard/gotth-extensions v1.0.0 h1:lKXOydN/7+Mq9W2juqnWUXNJiirpYR6F3KPXYV55Hho=' go.sum
grep -Fqx 'github.com/gotthboard/gotth-sdk v1.0.0 h1:RmjRs/l+H3KMDmnimZs2OL8ExCchT/opVO5tDdjb4EA=' go.sum
cache=$GOMODCACHE
source=proto/gotth/extensions/v1/control.proto
source_sha=846dbdd487e9688a338c7885478e1296456f994813330ecccdc3f4b225f10cfe
descriptor_sha=62a549b29a030f0decb46b2b1c2f382bbbc75430ba0b3d7c729bd21528b505cf
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT HUP INT TERM
printf '%s  %s\n' "$source_sha" "$cache/github.com/gotthboard/gotth-extensions@v1.0.0/$source" > "$tmp/extensions.sha256"
sha256sum --check --status "$tmp/extensions.sha256"
printf '%s  %s\n' "$source_sha" "$cache/github.com/gotthboard/gotth-sdk@v1.0.0/$source" > "$tmp/sdk.sha256"
sha256sum --check --status "$tmp/sdk.sha256"
cmp "$cache/github.com/gotthboard/gotth-extensions@v1.0.0/$source" "$cache/github.com/gotthboard/gotth-sdk@v1.0.0/$source"
[ "$(cat "$cache/github.com/gotthboard/gotth-extensions@v1.0.0/compat/v1/control-descriptor.sha256")" = "$descriptor_sha" ]
[ -z "$(find . -type f -path '*/gotth/extensions/v1/control*' -print -quit)" ] || { echo 'repository-local control copy is forbidden' >&2; exit 1; }
[ ! -e vendor ] || { echo 'repository-local dependency source is forbidden' >&2; exit 1; }
rg_reject_matches /usr/bin/rg 'repository-local control import is forbidden' -n 'gotth-extension-markdown/proto/gotth/extensions/v1' --glob '*.go' .
printf '%s\n' 'published foundation/SDK binding verified; no local control copy'
