# Product requirements

Status: initial product scope only; no implementation, module, integration, or
release exists, and implementation and release acceptance remain pending.

## Objective

Provide one reusable GOTTH rendering extension that combines Markdown/GFM,
Mermaid, and math without requiring each product to maintain separate engines.

## Requirements

- **MD-001 — Unified packaging:** one extension identity and deployable
  container; Mermaid and math are internal capabilities, not separate extensions.
- **MD-002 — Compatibility:** provide a pinned BB-compatible Markdown profile.
  Reuse appropriate GOTTH-BB parsing, sanitization and regression work only after
  separating forum-specific dependencies. Existing source, rendering, search
  projection and migration contracts must be preserved or explicitly migrated.
- **MD-003 — Opt-in rich rendering:** profiles explicitly enable Mermaid and
  inline/display math. Merely upgrading or installing the extension must not
  activate new syntax for existing content.
- **MD-004 — Repeatable output:** bind results to exact engine, profile,
  sanitizer, configuration and asset identities. Define and verify deterministic
  behavior before enabling persistence or render caches.
- **MD-005 — Authority:** hosts own content, authentication, permissions,
  persistence and publication policy. The renderer has no product-database or
  container-engine access and no runtime outbound resource fetching.
- **MD-006 — Safe complete results:** reject invalid, incomplete or oversized
  results. Enforce explicit input/output, time, memory and concurrency limits;
  return structured errors without logging private source content.
- **MD-007 — Host acceptance:** returned output is untrusted until validated by
  the host's approved HTML/SVG/MathML and URL policy. Rich output must not bypass
  the current forum policy or silently broaden it.
- **MD-008 — Availability:** host-retained accepted content/assets remain
  readable when the extension is unavailable, subject to revocation policy.
  Preview and new publication report rendering unavailability, with no silent
  fallback to a different renderer.
- **MD-009 — Presentation:** published output requires no author script or
  third-party CDN. Bundle required styles/fonts and define accessible equation
  and diagram representations during specification.
- **MD-010 — Published control foundation:** any future host/control adoption
  must select the exact published `gotth-extensions v1.0.0` and
  `gotth-sdk v1.0.0` modules. Manifest, grant, session, negotiation, and
  lifecycle behavior comes from
  `github.com/gotthboard/gotth-extensions/pkg/extensions`; generated
  control messages and services come from
  `github.com/gotthboard/gotth-sdk/gen/go/gotth/extensions/v1`.
- **MD-011 — Source ownership:** import the generated V1 control package from
  the SDK. Do not copy or regenerate the control protobuf or generated bindings
  in this repository. The generic control service remains handshake and health
  only; the Markdown rendering interface is separately specified, extension-provided,
  host-called business interface and cannot acquire host authority by being
  adjacent to the control service.

## Non-goals

This extension is not an editor, drag-and-drop layout engine, application
workflow designer, identity provider, media uploader or general code executor.
Footnotes, emoji shortcodes, mentions, issue references, arbitrary embeds and
syntax highlighting are not implicitly included by this scope.

## Acceptance direction

Implementation requires a concrete interface and migration specification, then
behavioral compatibility, security/boundary, rendering, accessibility, lifecycle,
failure and performance evidence. The roadmap tracks these as future work;
none is claimed complete by creating this repository or recording exact future
dependency pins. A module, generated tree, framework skeleton, or release-shaped
scaffolding is not acceptance evidence and must not be added solely to make the
design appear adopted.
