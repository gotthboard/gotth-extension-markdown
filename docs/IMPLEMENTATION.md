# Markdown extension v1.0.0 implementation specification

Status: implemented source candidate on `workflow/feature/v1-release`; not tagged,
published, deployed, or integrated into a host.

## V1 product boundary

V1 contains one extension, `gotth.render.markdown`, and one profile,
`gotth.markdown.bb-gfm.v1`. The profile is Markdown/GFM compatible with the
GOTTH-BB Alpha.3 renderer: Goldmark v1.8.5 with linkify (HTTP/HTTPS only),
tables without alignment attributes, strikethrough, and disabled task-list
checkboxes, followed by the BB p2 Bluemonday v1.0.27 allowlist.

Mermaid, inline/display math, author HTML, syntax highlighting, footnotes,
mentions, embeds, scripts, fonts, and browser rendering are not V1 capabilities.
A fenced block labelled `mermaid` remains an ordinary sanitized code block;
math delimiters remain text. Adding a rich profile requires a separate versioned
contract and acceptance record. Installing V1 cannot reinterpret existing
content.

## Published foundation and control

The module selects exactly `github.com/gotthboard/gotth-extensions v1.0.0`
and `github.com/gotthboard/gotth-sdk v1.0.0`, without replacement. Manifest,
digest, and validation behavior is imported from
`github.com/gotthboard/gotth-extensions/pkg/extensions`. The process imports
and registers the SDK-generated
`github.com/gotthboard/gotth-sdk/gen/go/gotth/extensions/v1` control service.
There is no repository-local control protobuf or generated control package.

`gotth.extensions.v1.ExtensionControl` remains handshake and health only.
The extension-owned host-called rendering service is
`gotth.markdown.renderer.v1.MarkdownRenderer`, generated from
`proto/gotth/markdown/renderer/v1/renderer.proto`. Sharing one authenticated
Unix-socket gRPC server does not merge control and business authority.

The manifest exposes control 1.0, renderer interface 1.0, and capability
`markdown.render`; it requests no secrets. Startup byte-exactly parses the
canonical foundation grant and calls foundation `Negotiate` against the exact
control-1.0/renderer-1.0 host profile. It retains the complete negotiated session;
handshake advertises only that session, health requires its fingerprint, and
rendering requires both its exact interface and capability.

## Runtime files and local transport

The executable accepts paths, never credential values:

- `GOTTH_EXTENSION_CONFIG_FILE`: strict private configuration JSON;
- `GOTTH_EXTENSION_GRANT_FILE`: canonical `gotth.extensions.grant.v1` authority grant;
- `GOTTH_EXTENSION_SERVICE_TOKEN_FILE`: local RPC token;
- `GOTTH_EXTENSION_SOCKET`: absent socket path in a private runtime directory.

Files must be canonical absolute, regular, non-symlinked, owned by the effective
UID, and have no group/other permission bits. Files are checked before and after
open and are bounded to 64 KiB; the token file is bounded to 8 KiB and must
contain exactly 32–256 bytes, each in printable ASCII `0x21..0x7e`. The socket parent must be canonical, private,
owner-controlled, and the socket path must not preexist. The process sets socket
mode `0600`.

Configuration has exactly four fields:

```json
{"schema":"gotth.extension.markdown.config.v1","profile":"gotth.markdown.bb-gfm.v1","timeout_milliseconds":500,"max_concurrent":4}
```

Timeout is 50–2000 ms and concurrency is 1–8. The displayed single-line JSON is
the canonical byte encoding: exact field order, JSON spelling, and no surrounding
whitespace or newline. Unknown, duplicate, missing, trailing, reordered,
escaped-equivalent, or otherwise noncanonical bytes reject startup. The grant must be byte-for-byte
`CanonicalGrant`, bind the linked manifest digest and extension ID, select exactly
renderer interface 1.0 and `markdown.render`, and select no secrets. Foundation
negotiation derives the instance ID, manifest/grant digests, and session
fingerprint; callers cannot supply those derived values independently.

Every decoded RPC first passes a unary interceptor requiring exactly one
`x-gotth-service-token` metadata value whose fixed-size SHA-256 digest is
compared in constant time. Only authenticated calls reach application deadline,
correlation, binding, and request validation. Every RPC also requires a caller
deadline and lowercase UUID correlation ID. Rendering also requires the bound
capability. No request or error log contains source,
HTML, visible text, token, or runtime-file contents.

SIGINT/SIGTERM starts graceful gRPC shutdown; a two-second ceiling forces stop.
The process has a 128 MiB Go memory limit, gRPC message limits, and a
process-global nonblocking render semaphore. Once all configured render slots are
occupied, additional calls from any connection fail with `RESOURCE_EXHAUSTED`
rather than waiting in an unbounded queue. The host/container still applies an
independent hard memory and CPU limit. V1 starts no subprocess and embeds no browser.

## Render request and complete result

`RenderRequest` contains only `correlation_id`, exact `profile`, and
`source`. Source must be nonempty, non-whitespace valid UTF-8 of at most 65,536
bytes. Document text cannot choose an engine, policy, URL fetch, timeout,
capability, asset, file, or subprocess.

A successful `RenderResponse` has:

- schema `gotth.markdown.render-response.v1` and `complete=true`;
- sanitized HTML, 1–262,144 bytes;
- BB p2-compatible NFC/Unicode-15 normalized visible text, 0–65,536 bytes; semantic documents such as a thematic break may intentionally project empty search text;
- extension version, engine, profile, sanitizer, configuration digest, and
  renderer identity.

The renderer identity is SHA-256 over UTF-8 JSON with no trailing newline and
exact key order `schema`, `extension_version`, `engine`, `profile`,
`sanitizer`, `configuration_sha256`, `assets`, `maximum_source_bytes`,
`maximum_html_bytes`, `maximum_visible_text_bytes`. Values are encoded with
Go `encoding/json`; `assets` is `none`; numeric limits are decimal integers.
For version `1.0.0` and configuration digest 64 lowercase `a` characters, the
canonical bytes are:

```json
{"schema":"gotth.markdown.render-response.v1","extension_version":"1.0.0","engine":"goldmark-v1.8.5-gfm","profile":"gotth.markdown.bb-gfm.v1","sanitizer":"bluemonday-v1.0.27-bb-p2","configuration_sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","assets":"none","maximum_source_bytes":65536,"maximum_html_bytes":262144,"maximum_visible_text_bytes":65536}
```

Their SHA-256 is
`f5050a806c1f941b8a806ae816fd1fdf358267d60b745909cde54c5ec7499820`.
It is suitable for a host cache/persistence key only together with tenant and
host output-policy identity.
No partial result is returned. For successfully decoded protobuf requests,
invalid application input is `INVALID_ARGUMENT`; output expansion is
`RESOURCE_EXHAUSTED`; caller cancellation is `CANCELED`; internal/configured
deadline is `DEADLINE_EXCEEDED`; unexpected engine failure is a source-redacted
`INTERNAL` status. Malformed protobuf bytes, including an invalid UTF-8 encoding
of the protobuf `string` source field, are rejected by gRPC decoding before the
application handler and currently surface as `INTERNAL`; a raw-wire process
regression pins that transport behavior.

Context cancellation is checked before parse, after Goldmark, and after
sanitization/projection. A synchronized in-process gRPC test waits until the server
render function is entered, cancels the client context, and proves that the server
observes cancellation before asserting `CANCELED`; the compiled-process fixture
only checks an already-canceled client call and is not cited as server-observation
evidence. Render admission never waits. Goldmark/Bluemonday are
in-process bounded CPU work and cannot be interrupted inside a single library
call; the source/message/process bounds are therefore part of the timeout
contract, not optional tuning.

## Sanitization, URLs, and no egress

Goldmark's unsafe HTML and unsafe-link rendering remain disabled. Bluemonday
admits only BB p2 structural tags, exact disabled checkbox attributes, and
`href` on anchors. Links may be relative or use HTTP or HTTPS and gain
`rel="nofollow noreferrer"`. Script, style, SVG, MathML, event attributes,
forms, active inputs, images, iframe, object, data URLs, and arbitrary attributes
are removed. A prefilter admits only Goldmark's exact lowercase disabled task
checkbox shape.

The executable has no outbound client, dialing call, shell, plugin, process
execution, runtime URL, proxy use, or asset fetch. Production-source policy
checks enforce that narrow mechanism. Deploy with no network namespace
(`--network=none` or equivalent); the Unix socket is the only application
transport.

## Accessibility and host acceptance

Output keeps semantic headings, lists, blockquotes, code, links, tables, and
disabled task controls. Visible text is independently projected from sanitized
HTML with deterministic block boundaries and whitespace normalization; it can be
used for indexing and a non-markup representation. V1 has no visual-only diagram
or equation output, author script, animation, font, or CDN dependency.

The output is still untrusted at the host boundary. A host must validate the
claimed complete response, identities, limits, allowed profile, and its own
HTML/URL policy before persistence or presentation. The extension does not
produce a host trusted-output type and cannot publish content. A host may use a
stricter URL policy than the extension.

## BB compatibility and migration contract

A BB adapter remains host-owned. For each candidate source it must call the exact
V1 profile, compare extension HTML and visible/search text with BB fixtures,
apply BB's final policy, and persist source, admitted output, renderer identity,
configuration digest, and host-policy identity atomically. Existing rows remain
on their recorded renderer. There is no read-time fallback or automatic bulk
migration. Migration is an explicit host operation with before/after counts,
exact source preservation, rollback material, and mismatch rejection.

Accepted output and source remain in host storage, so reads do not require a
live extension. Preview and new publication fail visibly when the extension is
unavailable; they never switch renderer. This repository does not alter BB,
Mail, a product database, or publication policy.

## Distribution and supply chain

Dependencies are pinned by `go.mod` and `go.sum`; the repository retains exact
license texts but no dependency source or generated control copy. From a
pre-seeded verified module cache, release builds require exact
`go1.27.3-X:nodwarf5 linux/amd64`. A statically linked, source-reproduced
`sanitized-run` launcher clears the ambient environment before starting any dynamic
tool, creates a fresh private `GOCACHE`/`GOTMPDIR`, and fixes Go, Git, compiler,
CGO, locale, and PATH state. One explicit cache-verification gate runs with
`GODEBUG=gocacheverify=1`; other gates reuse only that invocation's fresh cache.
Dirty evidence uses a manifest-bound filesystem snapshot that excludes `.git`.
Stable construction first rejects dirty/index-flagged state, freezes the exact HEAD
object, rejects tracked `.gitattributes`, exports that immutable commit, builds only
from the export, then rechecks the same HEAD and clean eligibility before publication.
Git configuration, ignores, replacement refs, index layout, and shared worktree state
therefore cannot inject mutable bytes. Builds use `-mod=readonly`, `GOPROXY=off`,
`GOSUMDB=off`, `CGO_ENABLED=0`, `-trimpath`, an empty build ID, fixed
`linux/amd64`, sorted archive entries, fixed `0755` directory modes, canonical
epoch zero, numeric ownership, and explicit `gzip -n -9`. The declared host-input closure includes all GOROOT files, the enumerated
shell/construction/validation executables and resolved dynamic libraries, and the
observed race-only CGO/GCC assembler/linker/startup/header/library chain. It is
not represented as a syscall-trace proof that no unobserved host file was read;
cross-host byte identity and hostile drift tests provide separate evidence. The builder uses no unchecked pipelines, validates and extracts the
temporary archive, verifies its checksums/modes/version, and only then links it to
the final non-overwriting destination.
The release archive contains only the static executable, canonical manifest,
scratch `Containerfile`, project license, third-party notices and exact license
texts, SPDX 2.3 SBOM, and SHA-256 manifest. No foundation configuration-metadata
format exists in the pinned V1 contract, so the archive invents none.

The scratch image runs as UID/GID 65532 and contains the executable plus the
project license, third-party notices and exact license/notice tree, and SPDX SBOM
under `/usr/share`; an archive inspection test pins every `COPY` instruction and
the corresponding payload bytes. Runtime files and socket directory are mounts
owned by that identity. The host supplies
hard memory/CPU/process limits and no network. Stable archive creation refuses an existing destination and always requires a
clean source tree. `GOTTH_EXTENSION_EVIDENCE_DIRTY=1` is the sole dirty-tree
override and emits only a visibly suffixed `DIRTY-EVIDENCE` non-release archive.

## Performance acceptance

On the designated development host with the exact release toolchain, five
three-second benchmark samples must report both the representative GFM fixture
and a 65,536-byte source at no more than 100 ms/op and 16 MiB allocated/op. The
compiled-process fixture must expose its socket within three seconds and complete
handshake, health, an authenticated hostile-input render, transport-limit
rejection, an already-canceled client call, and graceful termination. Server-
observed cancellation is proved separately by the synchronized in-process gRPC
test. These are admission ceilings,
not general latency guarantees; consuming hosts set capacity below the configured
concurrency maximum and retain the 500 ms default request timeout unless their
own measurements justify another admitted value.

## Acceptance gates

V1 admission requires exact dependency/control binding, unit and compiled
Unix-process tests, hostile-input and output-expansion tests, cancellation and
timeout tests, race/repeat/shuffle/fuzz runs, vet/build/format/diff checks, pinned protobuf
regeneration parity, network-disabled cache-seeded build, exact parsed
dependency/version/license/SPDX relationship parity, reproducible archive, representative
performance results on the development host, two fresh independent reviews, and an exact
source/evidence seal. Tag, release publication, image publication, deployment,
and host integration are separate authorized actions.
