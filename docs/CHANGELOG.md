# Changelog

## Unreleased

### 2026-10-23 CDT — Bind the future control boundary to published V1

Affected files: README.md, docs/PRD.md, docs/ARCHITECTURE.md,
docs/ROADMAP.md, docs/CHANGELOG.md.

Require any future host/control implementation to pin the exact published
`gotth-extensions v1.0.0` and `gotth-sdk v1.0.0` modules. Foundation contracts
must be imported from `github.com/gotthboard/gotth-extensions/pkg/extensions`,
and generated control messages and services must be imported from
`github.com/gotthboard/gotth-sdk/gen/go/gotth/extensions/v1` rather than copied
or regenerated in this repository. Keep the generic control service limited to
handshake and health; the Markdown render business interface remains separate,
extension-provided, host-called, and unspecified.

Verification: documentation links, formatting/whitespace, changed-file scope,
and the absence of modules, generated code, executable scaffolding, commits,
tags, and releases. No runtime or performance result is claimed.

Status: documentation-only design update. There is still no implementation or
release. No module, integration candidate, tag, container, or deployment exists.
The published pins constrain future work; they do not make this repository adopted.

### 2026-10-02 CDT — Establish unified rendering extension

Commit: current commit; hash assigned by Git after commit

Affected files: README.md, LICENSE, SECURITY.md, docs/PRD.md,
docs/ARCHITECTURE.md, docs/ROADMAP.md, docs/CHANGELOG.md.

Record one extension/container for Markdown/GFM, Mermaid and math; define
initial product scope, proposed host/renderer boundaries and staged development.
The editor remains a separate consumer. Use the GOTTH MIT license convention.

Verification: documentation/link/whitespace inspection and independent review
before initial publication. No runtime test or performance result is claimed.

Non-goals: no extracted production code, installed engines, BB behavior change,
container implementation, release or deployment. Engine and interface decisions
remain future specification work.
