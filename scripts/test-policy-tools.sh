#!/bin/sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd -P)
. "$root/scripts/pinned-toolchain.sh"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT HUP INT TERM
. "$root/scripts/rg-policy-lib.sh"
for status in 0 1 2 17; do
  cat > "$tmp/rg-$status" <<INNER
#!/bin/sh
exit $status
INNER
  chmod 0755 "$tmp/rg-$status"
done
if rg_reject_matches "$tmp/rg-0" forbidden pattern >/dev/null 2>&1; then echo 'rg match was accepted' >&2; exit 1; fi
rg_reject_matches "$tmp/rg-1" forbidden pattern
for status in 2 17; do
  set +e; rg_reject_matches "$tmp/rg-$status" forbidden pattern >/dev/null 2>&1; got=$?; set -e
  [ "$got" -eq "$status" ] || { echo "rg status $status became $got" >&2; exit 1; }
done
set +e; rg_reject_matches "$tmp/missing-rg" forbidden pattern >/dev/null 2>&1; missing=$?; set -e
[ "$missing" -eq 127 ] || { echo "missing rg status became $missing" >&2; exit 1; }

sed 's#/usr/bin/rg d006acfe#/missing/rg d006acfe#' "$root/scripts/pinned-toolchain.sh" > "$tmp/missing-toolchain.sh"
if /bin/sh -c '. "$1"' sh "$tmp/missing-toolchain.sh" >/dev/null 2>&1; then echo 'missing pinned tool was accepted' >&2; exit 1; fi
sed 's#d006acfe51619144#0000000000000000#' "$root/scripts/pinned-toolchain.sh" > "$tmp/drifted-toolchain.sh"
if /bin/sh -c '. "$1"' sh "$tmp/drifted-toolchain.sh" >/dev/null 2>&1; then echo 'drifted pinned tool identity was accepted' >&2; exit 1; fi

for required in \
  /usr/bin/basename /usr/bin/env /usr/bin/mv /usr/bin/rmdir \
  /usr/bin/gcc /usr/bin/g++ /usr/bin/ar /usr/bin/as /usr/bin/ld \
  /usr/lib/go/bin/go /usr/lib/go/pkg/tool/linux_amd64/compile /usr/lib/go/pkg/tool/linux_amd64/link \
  /usr/lib/go/pkg/tool/linux_amd64/asm /usr/lib/go/pkg/tool/linux_amd64/cgo \
  /usr/lib/go/src/runtime/runtime.go /usr/lib/go/pkg/include/textflag.h \
  /usr/include/stdio.h /bin/sh /usr/bin/ldd /usr/bin/tar /usr/bin/gzip /usr/bin/sha256sum /usr/bin/protoc /usr/bin/rg
do
  grep -Fq "  $required" "$root/scripts/release-toolchain.sha256" || { echo "release toolchain manifest omitted required input: $required" >&2; exit 1; }
done
for required in crt1.o crti.o crtbeginS.o crtendS.o crtn.o libgcc.a libc.so; do
  grep -Eq "  .*/$required$" "$root/scripts/release-toolchain.sha256" || { echo "release toolchain manifest omitted GCC/link input: $required" >&2; exit 1; }
done
while IFS= read -r required; do
  [ -n "$required" ] || continue
  grep -Fq "  $required" "$root/scripts/release-toolchain.sha256" || { echo "observed race input omitted from closure: $required" >&2; exit 1; }
done < "$root/scripts/release-race-inputs.paths"

toolchain_root="$tmp/toolchain-root"
mkdir -p "$toolchain_root/scripts"
cp "$root/go.mod" "$toolchain_root/"
cp "$root/scripts/sanitized-run" "$root/scripts/sanitized-run.go" "$root/scripts/sanitized-run.sha256" "$root/scripts/pinned-toolchain.sh" \
   "$root/scripts/release-toolchain-paths.sh" "$root/scripts/release-toolchain-dynamic-paths.sh" \
   "$root/scripts/release-race-inputs.paths" "$root/scripts/release-toolchain-dynamic.paths" "$root/scripts/verify-release-toolchain.sh" \
   "$root/scripts/release-toolchain.sha256" "$toolchain_root/scripts/"
sed -i '\#  /usr/lib/go/pkg/tool/linux_amd64/compile$#s/^[0-9a-f]\{64\}/0000000000000000000000000000000000000000000000000000000000000000/' "$toolchain_root/scripts/release-toolchain.sha256"
if "$toolchain_root/scripts/verify-release-toolchain.sh" >/dev/null 2>&1; then echo 'drifted compiler identity in complete release closure was accepted' >&2; exit 1; fi
cp "$root/scripts/release-toolchain.sha256" "$toolchain_root/scripts/release-toolchain.sha256"
dynamic_library=$(head -n 1 "$root/scripts/release-toolchain-dynamic.paths")
grep -Fq "  $dynamic_library" "$root/scripts/release-toolchain.sha256" || { echo "dynamic library content omitted from toolchain hash manifest: $dynamic_library" >&2; exit 1; }
/usr/bin/awk -v p="$dynamic_library" 'BEGIN { z="0000000000000000000000000000000000000000000000000000000000000000" } { if ($2 == p) $1=z; print $1 "  " $2 }' "$toolchain_root/scripts/release-toolchain.sha256" > "$toolchain_root/scripts/release-toolchain.sha256.new"
mv "$toolchain_root/scripts/release-toolchain.sha256.new" "$toolchain_root/scripts/release-toolchain.sha256"
if "$toolchain_root/scripts/verify-release-toolchain.sh" >/dev/null 2>&1; then echo 'same-path dynamic-library byte drift was accepted' >&2; exit 1; fi
cp "$root/scripts/release-toolchain.sha256" "$toolchain_root/scripts/release-toolchain.sha256"
sed -i '1s#^.*$#/missing/dynamic-library#' "$toolchain_root/scripts/release-toolchain-dynamic.paths"
if "$toolchain_root/scripts/verify-release-toolchain.sh" >/dev/null 2>&1; then echo 'drifted effective dynamic-library closure was accepted' >&2; exit 1; fi

fixture="$tmp/policy-fixture"
mkdir -p "$fixture/scripts" "$fixture/cmd/gotth-extension-markdown" "$fixture/internal/x" "$fixture/proto/gotth/markdown/renderer/v1"
cp "$root/scripts/check-source-boundary.sh" "$root/scripts/pinned-toolchain.sh" "$root/scripts/rg-policy-lib.sh" "$fixture/scripts/"
for n in 1 2 3; do : > "$fixture/proto/gotth/markdown/renderer/v1/$n.proto"; done
cat > "$fixture/cmd/gotth-extension-markdown/main.go" <<'INNER'
package main
import "net"
func main() { _, _ = net.Listen("unix", runtimeConfig.SocketPath) }
var runtimeConfig struct{ SocketPath string }
INNER
: > "$fixture/internal/x/x.go"
"$fixture/scripts/check-source-boundary.sh" >/dev/null
for snippet in \
  'var _ = http.Get' \
  'var _ = os.StartProcess' \
  'var _ = syscall.Exec' \
  'var _ = net.ListenTCP' \
  'var _ = net.Listen("tcp", "127.0.0.1:0")' \
  'var _ = http.ListenAndServe'
do
  printf 'package x\n%s\n' "$snippet" > "$fixture/internal/x/x.go"
  if "$fixture/scripts/check-source-boundary.sh" >/dev/null 2>&1; then echo "source boundary accepted forbidden mechanism: $snippet" >&2; exit 1; fi
done
: > "$fixture/internal/x/x.go"

cat > "$tmp/preload.c" <<EOF_C
#include <fcntl.h>
#include <unistd.h>
__attribute__((constructor)) static void mark(void) { int fd=open("$tmp/preload.marker",O_WRONLY|O_CREAT|O_APPEND,0600); if(fd>=0){write(fd,"x",1);close(fd);} }
EOF_C
cat > "$tmp/audit.c" <<EOF_C
#define _GNU_SOURCE
#include <fcntl.h>
#include <link.h>
#include <unistd.h>
unsigned int la_version(unsigned int v) { int fd=open("$tmp/audit.marker",O_WRONLY|O_CREAT|O_APPEND,0600); if(fd>=0){write(fd,"x",1);close(fd);} return v; }
EOF_C
/usr/bin/gcc -shared -fPIC -o "$tmp/preload.so" "$tmp/preload.c"
/usr/bin/gcc -shared -fPIC -o "$tmp/audit.so" "$tmp/audit.c"
LD_PRELOAD="$tmp/preload.so" /usr/bin/true
LD_AUDIT="$tmp/audit.so" /usr/bin/true
[ -s "$tmp/preload.marker" ] && [ -s "$tmp/audit.marker" ] || { echo 'real loader injections did not activate in positive control' >&2; exit 1; }
rm -f "$tmp/preload.marker" "$tmp/audit.marker"
LD_PRELOAD="$tmp/preload.so" LD_AUDIT="$tmp/audit.so" CC=/forbidden/cc CGO_CFLAGS=-DHOSTILE \
  GIT_CONFIG_COUNT=1 GIT_CONFIG_KEY_0=core.attributesFile GIT_CONFIG_VALUE_0=/forbidden/attributes \
  "$root/scripts/sanitized-run" /bin/sh -c 'test -z "${LD_PRELOAD+x}${LD_AUDIT+x}${GIT_CONFIG_COUNT+x}"; test "$CC" = /usr/bin/gcc; test "$CGO_CFLAGS" = "-O2 -g"'
[ ! -e "$tmp/preload.marker" ] && [ ! -e "$tmp/audit.marker" ] || { echo 'loader injection crossed static sanitizer boundary' >&2; exit 1; }

PATH=/forbidden GOROOT=/forbidden GOFIPS140=v1.0 GOCACHE=/shared GOCACHEPROG=/forbidden/cache \
  CC=/forbidden/cc CXX=/forbidden/cxx AR=/forbidden/ar CGO_CFLAGS=-DHOSTILE CGO_LDFLAGS=-Wl,--hostile \
  LD_AUDIT=/forbidden/audit.so LD_PRELOAD=/forbidden/preload.so GIT_CONFIG_COUNT=1 \
  "$root/scripts/sanitized-run" "$root/scripts/verify-release-toolchain.sh" >/dev/null
printf '%s\n' 'scanner statuses, complete race/build closure, source-boundary fixtures, fresh cache, Git isolation, and real loader-injection boundary verified'
