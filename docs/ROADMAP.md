# Development roadmap

All executable work below is pending. These are workstreams, not release tags.

1. Specify the existing GOTTH extension-host contract, render interface,
   versioned profiles, failure semantics and BB compatibility/migration boundary.
2. Extract a narrowly reusable BB Markdown core and prove an end-to-end
   BB-compatible render slice through the extension adapter.
3. Add optional Mermaid rendering inside the same extension and qualify its
   isolation, assets, limits and output acceptance.
4. Add optional inline/display math inside the same extension and qualify its
   delimiters, fonts, accessibility and output acceptance.
5. Package one container with pinned dependencies and required license notices;
   prove lifecycle, cancellation, offline rendering and availability behavior.
6. Integrate explicit host profiles without silently changing existing content;
   complete compatibility, security, accessibility and performance review.

Each slice needs product-facing acceptance checks and bounded supporting work.
Do not build a general testing framework or qualify unrelated infrastructure.
Reuse established tools and completed evidence where applicable.

This repository creation does not migrate BB, schedule another agent, implement
the visual editor, authorize a release or deploy a container.
