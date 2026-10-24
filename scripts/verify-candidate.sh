#!/bin/sh
set -eu
PATH=/usr/lib/go/bin:/usr/bin:/bin; export PATH
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd -P)
. "$root/scripts/pinned-toolchain.sh"
cd "$root"
gate=${1:-all}
provided_manifest=${2:-}
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT HUP INT TERM
if [ -n "$provided_manifest" ]; then
  manifest=$(CDPATH= cd -- "$(dirname -- "$provided_manifest")" && pwd -P)/$(basename -- "$provided_manifest")
else
  manifest="$tmp/source-manifest.sha256"
  "$root/scripts/source-manifest.sh" "$manifest"
fi
"$root/scripts/verify-source-snapshot.sh" "$manifest"
manifest_sum=$(sha256sum "$manifest")
manifest_sum=${manifest_sum%% *}
printf 'bound_source_manifest=%s files=%s\n' "$manifest_sum" "$(wc -l < "$manifest")"

run_gate() {
  gate_name=$1
  shift
  "$root/scripts/verify-source-snapshot.sh" "$manifest"
  printf 'gate_start=%s\n' "$gate_name"
  "$@"
  "$root/scripts/verify-source-snapshot.sh" "$manifest"
  printf 'gate_pass=%s manifest=%s\n' "$gate_name" "$manifest_sum"
}

gate_format() { [ -z "$(gofmt -l cmd internal proto scripts test)" ]; }
gate_cache_verify() { env GODEBUG=gocacheverify=1 go test -run='^$' -count=1 ./internal/markdown; }
gate_process() {
  process_tmp=$(mktemp -d "$tmp/process.XXXXXX")
  go build -mod=readonly -trimpath -o "$process_tmp/gotth-extension-markdown" ./cmd/gotth-extension-markdown
  GOTTH_MARKDOWN_BINARY="$process_tmp/gotth-extension-markdown" go test -count=1 ./test/process
}
gate_diff() { "$root/scripts/check-candidate-diff.sh"; }

run_named() {
  case $1 in
    toolchain) run_gate toolchain "$root/scripts/verify-release-toolchain.sh" ;;
    cache) run_gate cache-verification gate_cache_verify ;;
    test) run_gate test go test -p=1 -count=1 ./... ;;
    race) run_gate race env CGO_ENABLED=1 go test -race -count=1 ./... ;;
    vet) run_gate vet go vet ./... ;;
    build) run_gate build go build -mod=readonly -trimpath ./... ;;
    format) run_gate format gate_format ;;
    binding) run_gate binding "$root/scripts/check-published-control.sh" ;;
    generated) run_gate generated "$root/scripts/check-generated.sh" ;;
    licenses) run_gate licenses "$root/scripts/check-dependencies.sh" ;;
    policy) run_gate source-policy "$root/scripts/check-source-boundary.sh"; run_gate policy-hostile-regressions "$root/scripts/test-policy-tools.sh"; run_gate source-manifest-hostile-regression "$root/scripts/test-source-manifest.sh" ;;
    process) run_gate process gate_process ;;
    artifact) run_gate artifact "$root/scripts/test-reproducible-artifact.sh" ;;
    offline) run_gate offline go build -mod=readonly -trimpath ./... ;;
    diff) run_gate diff gate_diff ;;
    *) echo "unknown verification gate: $1" >&2; exit 2 ;;
  esac
}

if [ "$gate" = all ]; then
  for item in toolchain cache test vet build format binding generated licenses policy process artifact offline diff; do run_named "$item"; done
else
  run_named "$gate"
fi
"$root/scripts/verify-source-snapshot.sh" "$manifest"
printf 'verification_complete manifest=%s\n' "$manifest_sum"
