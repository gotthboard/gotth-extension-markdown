# Product requirements

Status: v1.0.0 source candidate implemented; verification and release admission
are recorded separately. No tag, hosted release, deployment, or host migration
is claimed.

## Objective

Provide one reusable GOTTH rendering extension identity. V1 delivers a bounded,
deterministic BB-compatible Markdown/GFM profile. Mermaid and math remain future
opt-in profiles inside the same extension identity; they are not implemented or
implicitly enabled by V1.

## V1 requirements

- **MD-001 — Unified packaging:** one extension identity and one scratch-container
  recipe. Future Mermaid/math engines remain internal capabilities, not separate
  extensions.
- **MD-002 — Compatibility:** exact pinned BB-compatible GFM engine and sanitizer
  profile; host-owned fixture comparison and explicit migration contract.
- **MD-003 — Opt-in syntax:** V1 advertises only
  `gotth.markdown.bb-gfm.v1`. Mermaid fences remain code and math delimiters
  remain text.
- **MD-004 — Repeatable output:** response binds exact engine, profile,
  sanitizer, limits, configuration, and asset identity; deterministic tests and
  archive reproduction are release gates.
- **MD-005 — Authority:** host owns content, authentication, authorization,
  persistence, tenant/cache isolation, publication, and final output policy.
  The renderer has no product database, container-engine, callback, or outbound
  resource access.
- **MD-006 — Safe complete results:** explicit source/HTML/text/message/time,
  memory, and concurrency bounds; cancellation; stable redacted status codes;
  never partial success.
- **MD-007 — Host acceptance:** renderer output is untrusted until independently
  validated against the host's allowed profile, identities, HTML/URL policy, and
  limits.
- **MD-008 — Availability:** hosts retain source and accepted output. Reads do
  not call the extension; new preview/publication fails explicitly when it is
  unavailable, without renderer fallback.
- **MD-009 — Presentation:** semantic HTML plus visible-text projection, no
  author script, third-party CDN, font, or V1 asset dependency.
- **MD-010 — Published foundation:** select exact
  `gotth-extensions v1.0.0` and `gotth-sdk v1.0.0` modules without replace.
- **MD-011 — Source ownership:** import SDK-generated V1 control bindings; do not
  copy/regenerate control files. Control remains handshake/health only; render
  is a separate extension-owned host-called interface.
- **MD-012 — Local lifecycle:** authenticated deadline-bearing gRPC over an
  owner-private `0600` Unix socket, strict private runtime files, bounded
  shutdown, and no source-bearing logs.
- **MD-013 — Supply chain:** pinned checksum-verified linked graph, retained licenses,
  SPDX 2.3 SBOM, offline static build, deterministic non-overwriting archive,
  checksums, and scratch container recipe.

## Non-goals

V1 is not an editor, layout/workflow system, identity provider, uploader,
general code executor, syntax highlighter, Mermaid renderer, math renderer, or
host migration. Footnotes, emoji shortcodes, mentions, issue references,
arbitrary HTML/embeds, browser execution, and runtime fetching are excluded.

## Acceptance

Traceability is MD requirement → architecture and implementation specification →
source/tests/scripts → revision-matched evidence. Required gates are defined in
`docs/IMPLEMENTATION.md` and `workflow.toml`. Source implementation does not
itself prove release, deployment, or host acceptance.
