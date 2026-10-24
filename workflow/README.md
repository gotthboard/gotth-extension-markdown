# Workflow

`workflow.toml` is canonical state for the v1.0.0 source candidate. Evidence is
stored outside the repository at
`.artifacts/gotth-extension-markdown-v1-release-20261023` and must be bound to
one exact source manifest with pre/post proof for every accepted gate before the
feature may move from `active` to `done`. Two fresh independent reviews of that
final source/evidence seal are mandatory; reviews of rejected predecessors do not count.

The implementation unit includes only the Markdown/GFM profile. Mermaid, math,
host integration, commit, tag, release publication, image publication, and
deployment are not child features hidden under this state.
