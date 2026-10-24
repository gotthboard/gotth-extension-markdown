# Architecture

Status: v1.0.0 source candidate implemented; not released or deployed.

## Boundary

One Go process serves two separately registered gRPC services on one private
Unix socket:

1. SDK-owned `gotth.extensions.v1.ExtensionControl` for handshake and health;
2. extension-owned `gotth.markdown.renderer.v1.MarkdownRenderer` for render.

Both require caller deadline, constant-time local service-token authentication,
and UUID correlation. The renderer additionally requires the bound
`markdown.render` capability. Foundation manifest/digest contracts come from
`gotth-extensions v1.0.0`; generated control messages/stubs come from
`gotth-sdk v1.0.0`. No control source is copied here.

## Data flow

The host selects the sole V1 profile and sends bounded UTF-8 source. A bounded
semaphore admits work. Goldmark v1.8.5 parses GFM locally; a checkbox prefilter
and Bluemonday v1.0.27 apply the BB p2 sanitizer; an HTML tokenizer derives
BB p2-compatible NFC/Unicode-15 visible text. The service checks context between stages and returns one complete
result plus exact identities. Any stage failure returns a redacted gRPC status
and no partial document.

There is no file access after startup, outbound connection, callback,
subprocess, browser, remote asset, database, persistence, publication, cache, or
fallback path. Runtime and protocol limits bound input, output, concurrency,
time, messages, and Go memory. The deployment supplies independent hard cgroup
limits and no network.

## Trust and persistence

Source remains host-authoritative. Returned HTML/text is untrusted until the
host checks completeness, profile, identities, limits, and its own output
policy. The host stores accepted source/output/identity together and keys any
cache by tenant and host policy as well as renderer identity. Existing stored
content remains readable without this process. Revocation, migration, and
re-render are explicit host actions.

## Profiles and assets

V1 exposes only `gotth.markdown.bb-gfm.v1`; asset identity is `none`.
Mermaid and math remain separate future slices requiring their own syntax,
isolation, accessibility, deterministic asset, sanitizer, licensing, and
performance decisions. They cannot enter the BB profile.

## Distribution

The module graph is pinned by `go.mod`/`go.sum`; dependency source and generated
control code are not copied here. A verified pre-seeded module cache supports an
offline static build that produces one binary. A reproducible archive adds the
canonical manifest, licenses/notices (including gRPC `NOTICE.txt`), SPDX SBOM,
checksums, and a scratch/non-root Containerfile that copies the legal/SBOM payload
alongside the executable.
The archive is a release input, not evidence of publication.

Detailed wire, failure, migration, lifecycle, sanitizer, and acceptance
contracts are in [the V1 implementation specification](IMPLEMENTATION.md).
