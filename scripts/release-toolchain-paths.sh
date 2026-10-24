#!/bin/sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd -P)
. "$root/scripts/pinned-toolchain.sh"
[ "$#" -eq 0 ] || { echo "usage: $0" >&2; exit 2; }
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT HUP INT TERM
: > "$tmp/paths"
add_tree() {
  tree=$1
  [ -d "$tree" ] || { echo "release toolchain tree unavailable: $tree" >&2; exit 1; }
  find "$tree" -type f -print0 >> "$tmp/paths"
  find "$tree" -type l -print0 | while IFS= read -r -d '' link; do
    resolved=$(readlink -f "$link")
    if [ -f "$resolved" ]; then printf '%s\0' "$resolved"; fi
  done >> "$tmp/paths"
}
add_file() {
  path=$1
  [ -f "$path" ] || { echo "release toolchain input unavailable: $path" >&2; exit 1; }
  printf '%s\0' "$path" >> "$tmp/paths"
  resolved=$(readlink -f "$path")
  if [ "$resolved" != "$path" ]; then printf '%s\0' "$resolved"; fi >> "$tmp/paths"
}
add_tree /usr/lib/go
add_tree /usr/lib/gcc
while IFS= read -r path; do
  case $path in ''|'#'*) continue;; esac
  add_file "$path"
done < "$root/scripts/release-race-inputs.paths"
while IFS= read -r path; do add_file "$path"; done <<'TOOLS'
/bin/sh
/home/linus/go/bin/protoc-gen-go
/home/linus/go/bin/protoc-gen-go-grpc
/usr/bin/ar
/usr/bin/as
/usr/bin/awk
/usr/bin/basename
/usr/bin/bash
/usr/bin/cat
/usr/bin/chmod
/usr/bin/cmp
/usr/bin/cp
/usr/bin/cut
/usr/bin/date
/usr/bin/diff
/usr/bin/dirname
/usr/bin/env
/usr/bin/file
/usr/bin/find
/usr/bin/g++
/usr/bin/gcc
/usr/bin/git
/usr/bin/grep
/usr/bin/gzip
/usr/bin/head
/usr/bin/ip
/usr/bin/ld
/usr/bin/ldd
/usr/bin/ln
/usr/bin/mkdir
/usr/bin/mktemp
/usr/bin/mv
/usr/bin/pkg-config
/usr/bin/protoc
/usr/bin/readlink
/usr/bin/rg
/usr/bin/rm
/usr/bin/rmdir
/usr/bin/sed
/usr/bin/sha256sum
/usr/bin/sort
/usr/bin/stat
/usr/bin/sync
/usr/bin/tail
/usr/bin/tar
/usr/bin/tee
/usr/bin/timeout
/usr/bin/true
/usr/bin/uname
/usr/bin/unshare
/usr/bin/wc
/usr/bin/xargs
/usr/lib/ld-linux-x86-64.so.2
/usr/share/licenses/go/LICENSE
TOOLS
for input in cc1 cc1plus collect2 lto1 lto-wrapper liblto_plugin.so crt1.o Scrt1.o crti.o crtbegin.o crtbeginS.o crtend.o crtendS.o crtn.o libgcc.a libgcc_eh.a libgcc_s.so libc.so libc_nonshared.a; do
  case $input in cc1|cc1plus|collect2|lto1|lto-wrapper) path=$(/usr/bin/gcc -print-prog-name="$input") ;; *) path=$(/usr/bin/gcc -print-file-name="$input") ;; esac
  case $path in "$input"|'') echo "GCC input unavailable: $input" >&2; exit 1;; esac
  add_file "$path"
done
LC_ALL=C sort -zu "$tmp/paths"
