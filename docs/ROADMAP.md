# Development roadmap

All executable work below is pending. These are workstreams, not release tags;
there is no implementation, module, integration candidate, or release.

1. Specify the render business interface, versioned profiles, failure semantics
   and BB compatibility/migration boundary on top of the exact published
   `gotth-extensions v1.0.0` and `gotth-sdk v1.0.0` host/control boundary.
   Confirm that the generic control service remains handshake and health only.
2. Extract a narrowly reusable BB Markdown core and prove an end-to-end
   BB-compatible render slice through the extension adapter. When executable
   work is admitted, import foundation contracts from
   `github.com/gotthboard/gotth-extensions/pkg/extensions` and generated control
   code from `github.com/gotthboard/gotth-sdk/gen/go/gotth/extensions/v1`; do not
   copy or regenerate the control schema or generated bindings here.
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

Do not add a Go module, generated-code tree, adapter framework, or release
scaffolding merely to mark V1 adoption. The published version pins are design
inputs until an approved implementation slice imports and verifies them in real
host/control behavior.

This repository creation does not migrate BB, schedule another agent, implement
the visual editor, authorize a release or deploy a container. This documentation
update likewise does not implement, integrate, tag, publish, or release anything.
