# Distribution and release

No release action is performed by these instructions. Commit, tag, hosted
release, image publication, deployment, and host integration require separate
authorization.

## Deterministic archive

Seed and verify the Go module cache during a connected preparation phase, then
disable network access. Every accepted command below starts with the statically
linked sanitizer; calling the shell scripts directly fails closed.

```sh
go mod download
go mod verify
scripts/sanitized-run scripts/source-manifest.sh /external/evidence/source-manifest.sha256
# disable network here; GOPROXY=off alone is not a network-isolation proof
make verify SOURCE_MANIFEST=/external/evidence/source-manifest.sha256
scripts/sanitized-run scripts/build-artifact.sh 1.0.0 /empty/output
```

The output is `gotth-extension-markdown-1.0.0-linux-amd64.tar.gz`. The released
Extensions/SDK foundation records a historical builder image, but this extension
does not claim that image as runnable or reproduced: neither authorized host has
both the retained archive and permission to a container runtime.

The admitted host fallback is `scripts/release-toolchain.sha256`, SHA-256
`5f3de55cbf404842b947f038c317add2569278cfb2f2740053791ad2143473b6`.
Its 16,789 entries bind every regular GOROOT file, the 184 observed race
system inputs, and every GCC header, GCC internal programs and resources, assembler/linker startup objects and
libraries, and every command used by construction or validation—including
`basename`, `env`, `mv`, and `rmdir`. A separate 90-path manifest, SHA-256
`569a709b353bf85d03bf6796889eba71157f9f858d518d32d7ab8e1750d5a704`,
binds the effective dynamic-library closure. The verifier reconstructs both
manifests, checks the sealed sanitizer source/binary pair, confirms that the
binary has no dynamic interpreter, and rejects compiler or library drift. This is
a declared closure derived from all GOROOT files, the enumerated tools, resolved
dynamic libraries, and observed race inputs—not a syscall-trace claim that no
unobserved host file was read. The designated development gate rebuilds that
binary byte-for-byte from its source, and the archive reproduces across the two
authorized hosts.

The sanitizer clears the complete inherited environment before any dynamic tool
starts. It fixes PATH, locale, Go target/toolchain/workspace/module settings,
`CC`, `CXX`, `AR`, all admitted `CGO_*` flags, Git configuration and
replacement-ref state, and canonical epoch zero. Every invocation receives a
fresh mode-0700 `GOCACHE` and `GOTMPDIR`; one explicit gate verifies cache
entries with `GODEBUG=gocacheverify=1`. A regression first proves that real
preload and audit libraries inject into an ordinary dynamic program, then proves that neither can
cross the static sanitizer boundary.

Dirty-evidence construction enumerates the filesystem independently of Git, rejects
symlinks, binds every source byte in the supplied source manifest, and copies those
exact files into a temporary snapshot. Stable construction rejects dirty state,
assume-unchanged/skip-worktree flags, and tracked `.gitattributes`; freezes the exact
HEAD object; builds only from that immutable export; then rechecks unchanged HEAD
and clean eligibility before non-overwriting publication. Git configuration, ignores,
replacement refs, index layout, and shared worktree state cannot inject mutable bytes. Dirty evidence requires
`GOTTH_EXTENSION_EVIDENCE_DIRTY=1` and is visibly suffixed
`DIRTY-EVIDENCE`.

The builder refuses an existing destination, normalizes staged directory and file
modes, writes an intermediate tar without a pipeline, compresses with
`gzip -n -9`, extracts and compares the temporary archive, verifies checksums
and version, rechecks the source manifest, and only then publishes by
non-overwriting hard link. Reproducibility tests compare two fresh-cache builds
under hostile Go, compiler, Git, loader, archive, PATH, and umask inputs.
Extract and verify:

```sh
tar -xzf gotth-extension-markdown-1.0.0-linux-amd64.tar.gz
cd gotth-extension-markdown-1.0.0-linux-amd64
sha256sum -c SHA256SUMS
./gotth-extension-markdown --version
```

The archive contains the static binary, canonical manifest, project license,
third-party notices and licenses, SPDX 2.3 SBOM, scratch Containerfile, and
checksums. It contains no source, secret, control-schema copy, Mermaid/math engine,
browser, font, or remote asset.

## Container image

Build only from the verified extracted directory. The Containerfile is
`FROM scratch`, copies the binary plus project license, third-party notices and
license/notice tree, and SPDX SBOM into `/usr/share`, and runs as 65532:65532.
The artifact regression compares the complete Containerfile with the admitted
recipe and verifies each copied payload byte before image publication. Record the
image
digest separately; archive reproducibility does not prove a builder-specific OCI
manifest digest.

Required runtime constraints:

- read-only root filesystem;
- no network (`--network=none` or equivalent);
- private tmpfs/bind mount for runtime files and Unix socket, owned by 65532;
- only the three private files and absent socket path described in
  `docs/IMPLEMENTATION.md`;
- hard memory at or below the admitted host budget, CPU quota, PID limit, and no
  container-engine socket;
- no environment secret values.

## Release admission and rollback

Before an immutable tag/release, bind tests, two fresh independent reviews,
archive digest, SBOM,
source manifest, and release notes to one committed tree. Verify hosted tag,
asset checksum, and public/authenticated visibility independently after
publication on canonical Forgejo. GitHub is a separately maintained public
mirror and its parity is not a Forgejo release gate. A host rollout retains the prior image digest and stored content;
rollback changes only the selected process image/configuration. Existing content
is never bulk re-rendered by deployment.
