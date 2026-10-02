# gotth-extension-markdown

One planned GOTTH extension for **Markdown/GFM, Mermaid diagrams, and math**.

**Status: design-stage, documentation only.** No renderer executable, container
image, API implementation, release, or product integration is available yet.

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
