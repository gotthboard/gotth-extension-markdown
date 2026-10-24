#!/bin/sh
set -eu
if [ "$#" -ne 2 ]; then echo "usage: $0 <version> <output-directory>" >&2; exit 2; fi
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd -P)
. "$root/scripts/pinned-toolchain.sh"
"$root/scripts/verify-release-toolchain.sh"
version=$1
output_input=$2
[ "$version" = 1.0.0 ] || { echo "only the admitted v1.0.0 artifact may be built" >&2; exit 2; }
mkdir -p "$output_input"
output=$(CDPATH= cd -- "$output_input" && pwd -P)
case $output in "$root"|"$root"/*) echo "artifact output must be outside the source tree" >&2; exit 2;; esac
[ "$SOURCE_DATE_EPOCH" = 0 ] || { echo "canonical release epoch drifted" >&2; exit 1; }
case $GOTTH_EXTENSION_EVIDENCE_DIRTY in 0|1) ;; *) echo "GOTTH_EXTENSION_EVIDENCE_DIRTY must be 0 or 1" >&2; exit 2;; esac
head_commit=
if [ "$GOTTH_EXTENSION_EVIDENCE_DIRTY" = 0 ]; then
  "$root/scripts/check-committed-source.sh"
  head_commit=$(git -C "$root" rev-parse HEAD)
  if git -C "$root" ls-tree -r --name-only "$head_commit" | grep -Eq '(^|/)\.gitattributes$'; then
    echo "stable release source forbids tracked .gitattributes" >&2
    exit 1
  fi
fi
work=$(mktemp -d "$output/.gotth-extension-markdown.XXXXXX")
trap 'rm -rf "$work"' EXIT HUP INT TERM
manifest="$work/source-manifest.sha256"
mkdir "$work/source"
if [ "$GOTTH_EXTENSION_EVIDENCE_DIRTY" = 0 ]; then
  "$root/scripts/export-head-snapshot.sh" "$head_commit" "$work/source"
else
  "$root/scripts/source-manifest.sh" "$manifest"
  "$root/scripts/verify-source-snapshot.sh" "$manifest"
  find "$root" -path "$root/.git" -prune -o -type f -printf '%P\0' | LC_ALL=C sort -z > "$work/source-paths"
  tar -C "$root" --null --verbatim-files-from -T "$work/source-paths" -cf "$work/source.tar"
  tar --same-permissions -xf "$work/source.tar" -C "$work/source"
fi
source=$(CDPATH= cd -- "$work/source" && pwd -P)
if [ "$GOTTH_EXTENSION_EVIDENCE_DIRTY" = 0 ]; then
  "$source/scripts/source-manifest.sh" "$manifest"
fi
"$source/scripts/verify-source-snapshot.sh" "$manifest"
name=gotth-extension-markdown-$version-linux-amd64
if [ "$GOTTH_EXTENSION_EVIDENCE_DIRTY" = 1 ]; then
  name=$name-DIRTY-EVIDENCE
  echo "building non-release manifest-bound dirty-source evidence archive" >&2
fi
"$source/scripts/check-published-control.sh"
"$source/scripts/check-dependencies.sh"
"$source/scripts/check-generated.sh"
"$source/scripts/check-source-boundary.sh"
archive="$output/$name.tar.gz"
[ ! -e "$archive" ] || { echo "release artifact already exists" >&2; exit 1; }
stage="$work/stage"
mkdir -p "$stage/$name"
(
  cd "$source"
  go mod verify
  go build -mod=readonly -trimpath -buildvcs=false -ldflags "-s -w -buildid= -X github.com/gotthboard/gotth-extension-markdown/internal/markdown.BuildVersion=$version" -o "$stage/$name/gotth-extension-markdown" ./cmd/gotth-extension-markdown
  go run -mod=readonly ./cmd/gotth-extension-markdown-manifest "$version" > "$stage/$name/manifest.json"
)
[ "$("$stage/$name/gotth-extension-markdown" --version)" = "$version" ]
cp "$source/Containerfile" "$source/LICENSE" "$source/THIRD_PARTY_NOTICES.md" "$source/SBOM.spdx.json" "$stage/$name/"
cp -R "$source/third_party/licenses" "$stage/$name/THIRD_PARTY_LICENSES"
(
  cd "$stage/$name"
  find . -type f ! -name SHA256SUMS -printf '%P\0' > "$work/checksum-paths.unsorted"
  LC_ALL=C sort -z < "$work/checksum-paths.unsorted" > "$work/checksum-paths"
  xargs -0 sha256sum < "$work/checksum-paths" > SHA256SUMS
)
find "$stage/$name" -type d -exec chmod 0755 {} +
chmod 0755 "$stage/$name/gotth-extension-markdown"
find "$stage/$name" -type f ! -name gotth-extension-markdown -exec chmod 0644 {} +
tar_file="$work/$name.tar"
temporary="$work/$name.tar.gz"
tar --format=ustar --sort=name --mtime="@$SOURCE_DATE_EPOCH" --owner=0 --group=0 --numeric-owner -C "$stage" -cf "$tar_file" "$name"
gzip -n -9 -c "$tar_file" > "$temporary"
gzip -t "$temporary"
tar -tzf "$temporary" > "$work/archive-list.txt"
mkdir "$work/validate"
tar --same-permissions -xzf "$temporary" -C "$work/validate"
diff -r "$stage/$name" "$work/validate/$name"
[ -z "$(find "$work/validate/$name" -type d ! -perm 0755 -print -quit)" ] || { echo "noncanonical archive directory mode" >&2; exit 1; }
[ "$(stat -c %a "$work/validate/$name/gotth-extension-markdown")" = 755 ]
[ "$("$work/validate/$name/gotth-extension-markdown" --version)" = "$version" ]
(cd "$work/validate/$name" && sha256sum -c SHA256SUMS)
"$root/scripts/verify-source-snapshot.sh" "$manifest"
if [ "$GOTTH_EXTENSION_EVIDENCE_DIRTY" = 0 ]; then
  [ "$(git -C "$root" rev-parse HEAD)" = "$head_commit" ] || { echo "stable release HEAD changed during construction" >&2; exit 1; }
  "$root/scripts/check-committed-source.sh"
fi
ln "$temporary" "$archive"
sync -d "$archive"
sync -d "$output"
archive_sum=$(sha256sum "$archive")
archive_sum=${archive_sum%% *}
printf '%s  %s\n' "$archive_sum" "$archive"
