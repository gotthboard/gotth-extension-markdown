#!/bin/sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd -P)
. "$root/scripts/pinned-toolchain.sh"
[ "$#" -eq 0 ] || { echo "usage: $0" >&2; exit 2; }
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT HUP INT TERM
: > "$tmp/ldd"
while IFS= read -r executable; do
  : > "$tmp/one"
  if /usr/bin/ldd "$executable" > "$tmp/one" 2>&1; then
    cat "$tmp/one" >> "$tmp/ldd"
  elif grep -Fq 'not a dynamic executable' "$tmp/one"; then
    :
  else
    echo "unable to resolve dynamic closure: $executable" >&2
    cat "$tmp/one" >&2
    exit 1
  fi
done <<'TOOLS'
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
TOOLS
for input in cc1 cc1plus collect2 lto1; do
  executable=$(/usr/bin/gcc -print-prog-name="$input")
  [ -x "$executable" ] || { echo "GCC executable unavailable: $input" >&2; exit 1; }
  /usr/bin/ldd "$executable" >> "$tmp/ldd" 2>&1
done
grep -oE '/[^[:space:]]+' "$tmp/ldd" > "$tmp/paths.raw"
: > "$tmp/paths"
while IFS= read -r path; do readlink -f "$path" >> "$tmp/paths"; done < "$tmp/paths.raw"
LC_ALL=C sort -u "$tmp/paths"
