# Sourced by every accepted build and verification script. Accepted entry
# points must be launched by the statically linked sanitized-run binary so no
# dynamic loader, shell startup file, Git config, or ambient Go setting is read
# before this boundary is established.
[ "${GOTTH_SANITIZED_ENV:-}" = 1 ] || {
  echo "accepted release commands require scripts/sanitized-run" >&2
  exit 1
}

expect_env() {
  name=$1
  want=$2
  eval "got=\${$name-}"
  [ "$got" = "$want" ] || { echo "sanitized environment drifted: $name" >&2; exit 1; }
}

expect_env PATH /usr/lib/go/bin:/usr/bin:/bin
expect_env LC_ALL C
expect_env TZ UTC
expect_env GO111MODULE on
expect_env GOENV off
expect_env GOTOOLCHAIN local
expect_env GOWORK off
expect_env GOFLAGS ''
expect_env GOAMD64 v1
expect_env GOEXPERIMENT ''
expect_env GOFIPS140 off
expect_env GOCACHEPROG ''
expect_env GOOS linux
expect_env GOARCH amd64
expect_env GOPROXY off
expect_env GOSUMDB off
expect_env GOROOT /usr/lib/go
expect_env GOPATH /home/linus/go
expect_env GOMODCACHE /home/linus/go/pkg/mod
expect_env GOTELEMETRY off
expect_env GOVCS off
expect_env GOPRIVATE ''
expect_env GONOPROXY ''
expect_env GONOSUMDB ''
expect_env GODEBUG ''
expect_env CC /usr/bin/gcc
expect_env CXX /usr/bin/g++
expect_env AR /usr/bin/ar
expect_env PKG_CONFIG /usr/bin/pkg-config
expect_env CGO_CFLAGS '-O2 -g'
expect_env CGO_CPPFLAGS ''
expect_env CGO_CXXFLAGS '-O2 -g'
expect_env CGO_FFLAGS '-O2 -g'
expect_env CGO_LDFLAGS '-O2 -g'
expect_env SOURCE_DATE_EPOCH 0
expect_env GIT_CONFIG_NOSYSTEM 1
expect_env GIT_CONFIG_SYSTEM /dev/null
expect_env GIT_CONFIG_GLOBAL /dev/null
expect_env GIT_ATTR_NOSYSTEM 1
expect_env GIT_NO_REPLACE_OBJECTS 1

case $GOCACHE in /tmp/gm.*/gocache) ;; *) echo "GOCACHE is not fresh and isolated" >&2; exit 1;; esac
case $GOTMPDIR in /tmp/gm.*/tmp) ;; *) echo "GOTMPDIR is not isolated" >&2; exit 1;; esac
[ -d "$GOCACHE" ] && [ "$(/usr/bin/stat -c %a "$GOCACHE")" = 700 ] || { echo "isolated GOCACHE boundary drifted" >&2; exit 1; }
for loader_name in LD_AUDIT LD_DEBUG LD_LIBRARY_PATH LD_PRELOAD; do
  eval "loader_set=\${$loader_name+x}"
  [ -z "$loader_set" ] || { echo "loader environment was not cleared: $loader_name" >&2; exit 1; }
done
for ambient_name in GIT_CONFIG_COUNT GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE CC_FOR_TARGET CXX_FOR_TARGET CGO_CFLAGS_ALLOW CGO_CFLAGS_DISALLOW CGO_CPPFLAGS_ALLOW CGO_CPPFLAGS_DISALLOW CGO_CXXFLAGS_ALLOW CGO_CXXFLAGS_DISALLOW CGO_FFLAGS_ALLOW CGO_FFLAGS_DISALLOW CGO_LDFLAGS_ALLOW CGO_LDFLAGS_DISALLOW; do
  eval "ambient_set=\${$ambient_name+x}"
  [ -z "$ambient_set" ] || { echo "ambient environment was not cleared: $ambient_name" >&2; exit 1; }
done

pinned_tool() {
  tool_path=$1
  tool_sha=$2
  [ -x "$tool_path" ] || { echo "pinned tool unavailable: $tool_path" >&2; exit 1; }
  actual_tool_sha=$(/usr/bin/sha256sum "$tool_path")
  actual_tool_sha=${actual_tool_sha%% *}
  [ "$actual_tool_sha" = "$tool_sha" ] || { echo "pinned tool identity mismatch: $tool_path" >&2; exit 1; }
}

pinned_tool /bin/sh c14c8f498f3be96f076ceb07d65dba584a86dca3c86b5dce65ed0f6cae73358d
pinned_tool /usr/lib/go/bin/go 21401c8d2dae2e6db0d756f7b9ae5a885ac6c14882cbbe521c6f49b184da8402
pinned_tool /usr/lib/go/bin/gofmt a8da8273e1d2190e1d0afcf78e14582933da33e80360c1c460150808d9804c59
pinned_tool /usr/bin/protoc 4b7137d2236700cd9a6b92898ccbe3dd869917912e05017591af811c8387bfbf
pinned_tool /home/linus/go/bin/protoc-gen-go 4be7d281ba025a9c3091ec3e2d43e11b87ea102c4617601792642faf8fd479da
pinned_tool /home/linus/go/bin/protoc-gen-go-grpc 42d84fd65a6fddaf21319a6ec7275f568007921d14860f2728056ee900d5a25c
pinned_tool /usr/bin/rg d006acfe51619144af4709e85945fe3b5a6f05038d356c4d8f752a33b4ecae42
pinned_tool /usr/bin/git 2912a2a32f408c9b8122e476251b32f4b35f4c3f0c15961686ae87ca456d384c
pinned_tool /usr/bin/tar 8bd961dfeee3543f158f550566a66e78c713a9ad5b88432bab93b63f2bb9347c
pinned_tool /usr/bin/gzip b1f9e749ee0809d6e90100218e6bc13cdf3f26f23c243d5f62aae39162297cb1
pinned_tool /usr/bin/sha256sum e0a612dbb527eccb83bdb1b5aacfc23a99e40054e1f60e63ce6be779eefc2ea6

[ "$(/usr/lib/go/bin/go version)" = "go version go1.27.3-X:nodwarf5 linux/amd64" ] || { echo "unexpected Go toolchain version" >&2; exit 1; }
[ "$(/usr/lib/go/bin/go env GOROOT)" = /usr/lib/go ] || { echo "GOROOT boundary drifted" >&2; exit 1; }
[ "$(/usr/lib/go/bin/go env GOFIPS140)" = off ] || { echo "GOFIPS140 boundary drifted" >&2; exit 1; }
[ -z "$(/usr/lib/go/bin/go env GOCACHEPROG)" ] || { echo "GOCACHEPROG boundary drifted" >&2; exit 1; }
[ "$(/usr/lib/go/bin/go env CC)" = /usr/bin/gcc ] || { echo "CC boundary drifted" >&2; exit 1; }
[ "$(/usr/lib/go/bin/go env CXX)" = /usr/bin/g++ ] || { echo "CXX boundary drifted" >&2; exit 1; }
