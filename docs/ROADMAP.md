# Development roadmap

## V1.0.0 Markdown profile — implemented source candidate

- [x] Exact published Extensions/SDK v1.0.0 control boundary, with no local copy.
- [x] Separate renderer v1 wire contract and manifest capability.
- [x] Deterministic bounded BB-compatible GFM render and visible-text projection.
- [x] Sanitization, no-egress source boundary, local authentication, deadlines,
  cancellation, concurrency, memory/message limits, and graceful lifecycle.
- [x] Pinned checksum-verified dependencies, notices, SPDX SBOM, cache-seeded/reproducible
  artifact path, scratch container recipe, and release documentation.
- [x] Revision-matched full development-host, isolated-fuzz, performance,
  compiled-process, dependency-fetch-disabled build, and artifact evidence.
- [x] Earlier cold reviews, including blocking closure review E and non-overriding
  clean review F, are recorded; every E finding is remediated in this candidate.
- [x] Fresh candidate-bound network-isolation and advisory govulncheck evidence.
- [ ] Two fresh independent post-remediation admission reviews and a clean-commit
  stable artifact. The current source/evidence seal is review input, not admission.
- [ ] Authorized commit, tag, hosted release, image publication, deployment, and
  host integration. These are separate actions and are not implied by source
  completion.

## Later profiles — not V1

1. Mermaid: settle engine/browser isolation, deterministic IDs/layout, SVG and
   URL policy, textual alternatives, local assets, licenses, and performance.
2. Math: settle delimiters, KaTeX/other engine selection, MathML/HTML policy,
   accessible speech/text, fonts, local assets, licenses, and performance.
3. Host adoption: implement and separately admit explicit profile selection and
   BB fixture/migration behavior without changing existing rows implicitly.

No later item should create a renderer framework. Add one profile only when its
product contract and acceptance fixtures are approved.
