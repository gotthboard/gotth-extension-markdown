#!/bin/sh
set -eu
PATH=/usr/lib/go/bin:/usr/bin:/bin; export PATH
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd -P)
. "$root/scripts/pinned-toolchain.sh"
cd "$root"
protoc_bin=/usr/bin/protoc
protoc_gen_go=/home/linus/go/bin/protoc-gen-go
protoc_gen_go_grpc=/home/linus/go/bin/protoc-gen-go-grpc
[ "$($protoc_bin --version)" = 'libprotoc 36.1' ] || { echo 'unexpected protoc version' >&2; exit 1; }
[ "$($protoc_gen_go --version)" = 'protoc-gen-go v1.36.11' ] || { echo 'unexpected protoc-gen-go version' >&2; exit 1; }
[ "$($protoc_gen_go_grpc --version)" = 'protoc-gen-go-grpc 1.6.2' ] || { echo 'unexpected protoc-gen-go-grpc version' >&2; exit 1; }
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT HUP INT TERM
PATH=/home/linus/go/bin:/usr/lib/go/bin:/usr/bin:/bin "$protoc_bin" -I . \
  --go_out="$tmp" --go_opt=paths=source_relative \
  --go-grpc_out="$tmp" --go-grpc_opt=paths=source_relative \
  proto/gotth/markdown/renderer/v1/renderer.proto
cmp proto/gotth/markdown/renderer/v1/renderer.pb.go "$tmp/proto/gotth/markdown/renderer/v1/renderer.pb.go"
cmp proto/gotth/markdown/renderer/v1/renderer_grpc.pb.go "$tmp/proto/gotth/markdown/renderer/v1/renderer_grpc.pb.go"
printf '%s\n' 'renderer protobuf source/generated parity verified with pinned toolchain'
