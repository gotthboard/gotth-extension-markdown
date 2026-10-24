# gotth-extension-markdown

Bounded server-side Markdown/GFM rendering for GOTTH hosts.

**Status:** v1.0.0 source candidate. It is implemented but not committed by this
work item, tagged, published, deployed, or integrated into a host.

## V1 behavior

- one extension identity: `gotth.render.markdown`;
- one explicit profile: `gotth.markdown.bb-gfm.v1`;
- BB-compatible Goldmark v1.8.5 GFM plus Bluemonday v1.0.27 p2 policy;
- complete sanitized HTML and normalized visible text;
- deterministic renderer/configuration identity;
- authenticated deadline-bearing gRPC on a private Unix socket;
- fixed input/output/concurrency/time/memory/message limits;
- no outbound network, subprocess, browser, CDN, runtime asset, or database;
- exact published `gotth-extensions v1.0.0` and `gotth-sdk v1.0.0`
  control boundary, imported rather than copied.

Mermaid and math are not V1 features. Mermaid fences remain code blocks and math
syntax remains ordinary text.

## Build and verify

Dependencies are pinned by `go.mod`/`go.sum`. After the module cache has been
seeded and verified, release builds run with `GOPROXY=off` and `GOSUMDB=off`; no
dependency source or generated control copy is stored in this repository.

```sh
scripts/sanitized-run scripts/source-manifest.sh /external/evidence/source-manifest.sha256
make verify SOURCE_MANIFEST=/external/evidence/source-manifest.sha256
# race/repeat/fuzz/performance remain designated-development-host gates
```

Every accepted local gate checks that exact fileset/content manifest before and
after execution. The artifact path also verifies the complete checked-in
release-toolchain fileset/content manifest, including all GOROOT files, the race
CGO/GCC assembler/linker/startup/header/library chain, and every used construction
or validation executable. Accepted commands start through the checked-in static
`sanitized-run` launcher, which creates a fresh private Go cache and clears loader,
Git, compiler, Go, and archive state before any dynamic tool starts.

Release artifacts use `scripts/sanitized-run scripts/build-artifact.sh 1.0.0 <empty-output-dir>`.
The script never overwrites an existing archive. See
[distribution](docs/DISTRIBUTION.md) and [implementation](docs/IMPLEMENTATION.md).

## Runtime

The host supplies strict private configuration, canonical foundation grant, and
token files plus an absent socket path through environment variables documented
in the implementation specification. A typical configuration is:

```json
{"schema":"gotth.extension.markdown.config.v1","profile":"gotth.markdown.bb-gfm.v1","timeout_milliseconds":500,"max_concurrent":4}
```

Deploy the scratch image as UID/GID 65532. It contains the executable plus the
project and third-party legal payload and SPDX SBOM; use a private runtime-
directory mount,
no network, read-only root, and hard memory/CPU/process limits. The host retains
source and accepted output and independently enforces final HTML/URL policy.

## Documents

- [Product requirements](docs/PRD.md)
- [Architecture](docs/ARCHITECTURE.md)
- [V1 implementation specification](docs/IMPLEMENTATION.md)
- [Roadmap](docs/ROADMAP.md)
- [Distribution and release](docs/DISTRIBUTION.md)
- [Verification](docs/VERIFICATION.md)
- [Changelog](docs/CHANGELOG.md)
- [Security policy](SECURITY.md)
- [SPDX SBOM](SBOM.spdx.json)
- [Third-party notices](THIRD_PARTY_NOTICES.md)

## Source and reports

[Forgejo](https://git.dannyhunn.com/gotthboard/gotth-extension-markdown) is the
private canonical source. The
[GitHub repository](https://github.com/gotthboard/gotth-extension-markdown) is a
public one-way mirror maintained separately; mirror parity is not a Forgejo
release gate. Use [GitHub Issues](https://github.com/gotthboard/gotth-extension-markdown/issues)
for non-security bugs. Report vulnerabilities through
[GitHub private vulnerability reporting](https://github.com/gotthboard/gotth-extension-markdown/security/advisories/new),
never a public issue or Forgejo ticket.

## License

[MIT](LICENSE). Third-party terms are retained as exact license texts under
`third_party/licenses/` (including the required gRPC `NOTICE.txt`) and listed in
[THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md); dependency source is not vendored.
