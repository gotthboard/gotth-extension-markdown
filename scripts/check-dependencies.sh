#!/bin/sh
set -eu
PATH=/usr/lib/go/bin:/usr/bin:/bin; export PATH
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd -P)
. "$root/scripts/pinned-toolchain.sh"
cd "$root"
cmp /usr/share/licenses/go/LICENSE third_party/licenses/go/LICENSE
expected=$(cat <<'EOF'
github.com/aymerick/douceur v0.2.0
github.com/gorilla/css v1.0.1
github.com/gotthboard/gotth-extension-markdown
github.com/gotthboard/gotth-extensions v1.0.0
github.com/gotthboard/gotth-sdk v1.0.0
github.com/microcosm-cc/bluemonday v1.0.27
github.com/yuin/goldmark v1.8.5
golang.org/x/net v0.60.0
golang.org/x/sys v0.48.0
golang.org/x/text v0.42.0
google.golang.org/genproto/googleapis/rpc v0.0.0-20260526163538-3dc84a4a5aaa
google.golang.org/grpc v1.83.2
google.golang.org/protobuf v1.36.11
EOF
)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT HUP INT TERM
go list -mod=readonly -deps -f '{{with .Module}}{{.Path}}{{with .Version}} {{.}}{{end}}{{end}}' ./cmd/gotth-extension-markdown > "$tmp/modules.raw"
sed '/^$/d' "$tmp/modules.raw" > "$tmp/modules.nonempty"
LC_ALL=C sort -u "$tmp/modules.nonempty" > "$tmp/modules"
actual=$(cat "$tmp/modules")
[ "$actual" = "$expected" ] || { printf 'linked module set drifted:\n%s\n' "$actual" >&2; exit 1; }
go mod verify
[ ! -e vendor ] || { echo 'repository-local dependency source is forbidden' >&2; exit 1; }
cache=$GOMODCACHE
cmp "$cache/github.com/aymerick/douceur@v0.2.0/LICENSE" third_party/licenses/github.com/aymerick/douceur/LICENSE
cmp "$cache/github.com/gorilla/css@v1.0.1/LICENSE" third_party/licenses/github.com/gorilla/css/LICENSE
cmp "$cache/github.com/gotthboard/gotth-extensions@v1.0.0/LICENSE" third_party/licenses/github.com/gotthboard/gotth-extensions/LICENSE
cmp "$cache/github.com/gotthboard/gotth-sdk@v1.0.0/LICENSE" third_party/licenses/github.com/gotthboard/gotth-sdk/LICENSE
cmp "$cache/github.com/microcosm-cc/bluemonday@v1.0.27/LICENSE.md" third_party/licenses/github.com/microcosm-cc/bluemonday/LICENSE
cmp "$cache/github.com/yuin/goldmark@v1.8.5/LICENSE" third_party/licenses/github.com/yuin/goldmark/LICENSE
cmp "$cache/golang.org/x/net@v0.60.0/LICENSE" third_party/licenses/golang.org/x/net/LICENSE
cmp "$cache/golang.org/x/net@v0.60.0/PATENTS" third_party/licenses/golang.org/x/net/PATENTS
cmp "$cache/golang.org/x/sys@v0.48.0/LICENSE" third_party/licenses/golang.org/x/sys/LICENSE
cmp "$cache/golang.org/x/sys@v0.48.0/PATENTS" third_party/licenses/golang.org/x/sys/PATENTS
cmp "$cache/golang.org/x/text@v0.42.0/LICENSE" third_party/licenses/golang.org/x/text/LICENSE
cmp "$cache/golang.org/x/text@v0.42.0/PATENTS" third_party/licenses/golang.org/x/text/PATENTS
cmp "$cache/google.golang.org/genproto/googleapis/rpc@v0.0.0-20260526163538-3dc84a4a5aaa/LICENSE" third_party/licenses/google.golang.org/genproto/googleapis/rpc/LICENSE
cmp "$cache/google.golang.org/grpc@v1.83.2/LICENSE" third_party/licenses/google.golang.org/grpc/LICENSE
cmp "$cache/google.golang.org/grpc@v1.83.2/NOTICE.txt" third_party/licenses/google.golang.org/grpc/NOTICE.txt
cmp "$cache/google.golang.org/protobuf@v1.36.11/LICENSE" third_party/licenses/google.golang.org/protobuf/LICENSE
cmp "$cache/google.golang.org/protobuf@v1.36.11/PATENTS" third_party/licenses/google.golang.org/protobuf/PATENTS
go run ./scripts/check-sbom.go SBOM.spdx.json < "$tmp/modules"
printf '%s\n' 'linked dependency checksums, retained licenses/notices, and exact SPDX parity verified'
