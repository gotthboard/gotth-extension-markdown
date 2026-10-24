#!/bin/sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd -P)
. "$root/scripts/pinned-toolchain.sh"
base=$(mktemp -d); extract=$(mktemp -d); canonical=$(mktemp)
trap 'rm -rf "$base" "$extract" "$canonical"' EXIT HUP INT TERM
left="$base/left"; right="$base/right"
name=gotth-extension-markdown-1.0.0-linux-amd64
dirty_flag=0
if ! "$root/scripts/check-committed-source.sh" >/dev/null 2>&1; then
  mkdir "$left"
  if GOTTH_EXTENSION_EVIDENCE_DIRTY=0 "$root/scripts/build-artifact.sh" 1.0.0 "$left" >/dev/null 2>&1; then
    echo "uncommitted stable artifact build did not fail closed" >&2
    exit 1
  fi
  rmdir "$left"
  dirty_flag=1
  name=$name-DIRTY-EVIDENCE
fi
(
  cd "$base"
  umask 022
  PATH=/forbidden GOROOT=/forbidden GOENV=/forbidden/goenv GOTOOLCHAIN=auto GOWORK=/forbidden/go.work \
    GOFLAGS='-tags=ambient_should_not_apply' GOAMD64=v4 GOEXPERIMENT=fieldtrack GOFIPS140=v1.0 \
    GOCACHE=/forbidden/cache GOCACHEPROG=/forbidden/cache CC=/forbidden/cc CXX=/forbidden/cxx AR=/forbidden/ar \
    CGO_CFLAGS=-DHOSTILE CGO_CPPFLAGS=-DHOSTILE CGO_CXXFLAGS=-DHOSTILE CGO_LDFLAGS=-Wl,--hostile \
    GIT_CONFIG_COUNT=1 GIT_CONFIG_KEY_0=core.attributesFile GIT_CONFIG_VALUE_0=/forbidden/attributes \
    GZIP=-1 TAR_OPTIONS='--wildcards --exclude=*' LD_AUDIT=/forbidden/audit.so LD_PRELOAD=/forbidden/preload.so \
    GOTTH_EXTENSION_EVIDENCE_DIRTY=$dirty_flag "$root/scripts/sanitized-run" "$root/scripts/build-artifact.sh" 1.0.0 left >/dev/null
)
(
  cd "$base"
  umask 077
  PATH=/also-forbidden GOROOT=/other GOENV=/other/goenv GOTOOLCHAIN=path GOWORK=/other/go.work \
    GOFLAGS='-mod=vendor' GOAMD64=v3 GOEXPERIMENT=arenas GOFIPS140=latest \
    GOCACHE=/other/cache GOCACHEPROG=/other/cache CC=/other/cc CXX=/other/cxx AR=/other/ar \
    CGO_CFLAGS=-DOTHER CGO_CPPFLAGS=-DOTHER CGO_CXXFLAGS=-DOTHER CGO_LDFLAGS=-Wl,--other \
    GIT_CONFIG_GLOBAL=/other/gitconfig GZIP=-9 TAR_OPTIONS='--no-recursion' LD_AUDIT=/other/audit.so LD_PRELOAD=/other/preload.so \
    GOTTH_EXTENSION_EVIDENCE_DIRTY=$dirty_flag "$root/scripts/sanitized-run" "$root/scripts/build-artifact.sh" 1.0.0 right >/dev/null
)
cmp "$left/$name.tar.gz" "$right/$name.tar.gz"
gzip -t "$left/$name.tar.gz"
gzip -dc "$left/$name.tar.gz" > "$base/archive.tar"
gzip -n -9 -c "$base/archive.tar" > "$base/recompressed.tar.gz"
cmp "$left/$name.tar.gz" "$base/recompressed.tar.gz"
tar --same-permissions -xzf "$left/$name.tar.gz" -C "$extract"
[ "$(stat -c %a "$extract/$name/gotth-extension-markdown")" = 755 ]
[ -z "$(find "$extract/$name" -type d ! -perm 0755 -print -quit)" ] || { echo "noncanonical archive directory mode" >&2; exit 1; }
[ "$("$extract/$name/gotth-extension-markdown" --version)" = 1.0.0 ]
(cd "$extract/$name" && sha256sum -c SHA256SUMS)
[ ! -e "$extract/$name/configuration-metadata.json" ]
(
  cd "$root"
  go run -mod=readonly ./cmd/gotth-extension-markdown-manifest 1.0.0 > "$canonical"
)
cmp "$canonical" "$extract/$name/manifest.json"
cmp "$root/LICENSE" "$extract/$name/LICENSE"
cmp "$root/THIRD_PARTY_NOTICES.md" "$extract/$name/THIRD_PARTY_NOTICES.md"
cmp "$root/SBOM.spdx.json" "$extract/$name/SBOM.spdx.json"
cmp "$root/third_party/licenses/google.golang.org/grpc/NOTICE.txt" "$extract/$name/THIRD_PARTY_LICENSES/google.golang.org/grpc/NOTICE.txt"
cat > "$base/expected.Containerfile" <<'INNER'
FROM scratch
COPY gotth-extension-markdown /gotth-extension-markdown
COPY LICENSE /usr/share/licenses/gotth-extension-markdown/LICENSE
COPY THIRD_PARTY_NOTICES.md /usr/share/licenses/gotth-extension-markdown/THIRD_PARTY_NOTICES.md
COPY THIRD_PARTY_LICENSES /usr/share/licenses/gotth-extension-markdown/THIRD_PARTY_LICENSES
COPY SBOM.spdx.json /usr/share/doc/gotth-extension-markdown/SBOM.spdx.json
USER 65532:65532
ENTRYPOINT ["/gotth-extension-markdown"]
INNER
cmp "$base/expected.Containerfile" "$extract/$name/Containerfile"
before=$(sha256sum "$right/$name.tar.gz"); before=${before%% *}
if GOTTH_EXTENSION_EVIDENCE_DIRTY=$dirty_flag "$root/scripts/sanitized-run" "$root/scripts/build-artifact.sh" 1.0.0 "$right" >/dev/null 2>&1; then echo "existing artifact replaced" >&2; exit 1; fi
after=$(sha256sum "$right/$name.tar.gz"); after=${after%% *}
[ "$before" = "$after" ]
printf '%s\n' 'artifact reproducibility, sanitized inputs, compression, pre-publication validation, legal payload, and container recipe verified'
