# gotth-extension-markdown

One planned GOTTH extension for **Markdown/GFM, Mermaid diagrams, and math**.

**Status: design-stage, documentation only.** No renderer executable, Go module,
container image, API implementation, release, or product integration exists.

The goal is a reusable rendering extension for GOTTH-BB and other GOTTH
products. Markdown, diagrams, and equations belong to one extension identity
and deployable container, with versioned profiles controlling enabled features.
The visual editor and drag-and-drop page builder remain separate consumers.

## Planned capabilities

- CommonMark/GFM rendering compatible with GOTTH-BB's current renderer.
- Optional server-rendered Mermaid diagrams.
- Optional inline and display mathematics.
- Validated output, visible text, explicit failures, and renderer/profile identity.
- Locally supplied rendering assets; no required third-party CDN.

Hosts retain source content, permissions, persistence, and final output-policy
enforcement. Installing an engine must not silently reinterpret existing posts
or enable additional syntax.

## Published foundation target

Any future implementation of the host/control boundary must pin the published
[`gotth-extensions v1.0.0`](https://git.dannyhunn.com/gotthboard/gotth-extensions/releases/tag/v1.0.0)
and [`gotth-sdk v1.0.0`](https://git.dannyhunn.com/gotthboard/gotth-sdk/releases/tag/v1.0.0)
modules. Foundation manifest, grant, session, negotiation, and lifecycle types
come from `github.com/gotthboard/gotth-extensions/pkg/extensions`. Generated
control messages and service stubs come from
`github.com/gotthboard/gotth-sdk/gen/go/gotth/extensions/v1`; this repository
must not copy the control protobuf or generated code.

That published control surface supplies handshake and health, not a Markdown
rendering API. The rendering business interface remains future specification
work and must remain extension-provided and host-called, with policy and final
admission retained by the host. These version pins are a design constraint,
not an implementation claim: this repository intentionally has no module or
framework added merely to appear adopted.

## Documents

- [Product requirements](docs/PRD.md)
- [Architecture outline](docs/ARCHITECTURE.md)
- [Development roadmap](docs/ROADMAP.md)
- [Changelog](docs/CHANGELOG.md)
- [Security reporting](SECURITY.md)

## Source and reports

[Forgejo](https://git.dannyhunn.com/gotthboard/gotth-extension-markdown) is the
private canonical source. This [GitHub repository](https://github.com/gotthboard/gotth-extension-markdown)
is its public one-way mirror.

Use [GitHub Issues](https://github.com/gotthboard/gotth-extension-markdown/issues)
for non-security bugs and feature requests. Report vulnerabilities through
[private vulnerability reporting](https://github.com/gotthboard/gotth-extension-markdown/security/advisories/new),
not public issues or Forgejo tickets.

## License

[MIT](LICENSE). Future incorporated dependencies must retain their respective
licenses and notices; this initial repository contains no extracted BB code or
bundled rendering engine.
