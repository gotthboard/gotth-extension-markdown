# Architecture outline

Status: proposed design; no implementation or release exists. This is not a
complete wire specification or runtime proof.

## Published V1 host/control boundary

Future implementation must bind its generic extension boundary to the exact
published `gotth-extensions v1.0.0` and `gotth-sdk v1.0.0` module versions. It
must import foundation manifest, grant, session, negotiation, and lifecycle
contracts from `github.com/gotthboard/gotth-extensions/pkg/extensions` and the
generated control messages and service stubs from
`github.com/gotthboard/gotth-sdk/gen/go/gotth/extensions/v1`.

The control protobuf and generated Go files remain owned by those published
foundation/SDK releases. This repository must not vendor, copy, or regenerate a
second control schema or generated tree. It currently contains no Go module,
generated package, implementation, or framework; the version binding is a
design decision rather than evidence of adoption or release readiness.

The published `gotth.extensions.v1.ExtensionControl` surface is limited to
handshake and health. It does not provide rendering, installation, configuration,
secret, process-control, or generic invocation operations. A future render
interface therefore requires its own reviewed wire specification. It remains
extension-provided and host-called and cannot turn a grant into a callback or
move authentication, authorization, persistence, publication, or output-policy
authority out of the host.

## One extension, several internal renderers

An extension adapter invokes an internal Markdown parser and optional Mermaid
and math renderers. Bounded worker processes may be used inside the extension
container; products manage one extension rather than one per syntax feature.

Goldmark is the existing BB parser and a reuse candidate. Mermaid's supported
server rendering path and an engine such as KaTeX are candidates for the other
capabilities. No additional engine version, browser runtime, transport or
dependency is selected or installed by this outline.

## Request and response responsibilities

A conceptual request contains source, an authorized profile/version and bounded
render options. The host selects the policy; document content cannot expand
permissions, fetch resources, override limits, or select arbitrary engines.

A conceptual response contains rendered content, an asset manifest if needed,
visible-text/search representation, engine/profile identity and explicit failure
information. Transport, schema, asset representation, delimiter/escaping rules
and error codes remain to be specified. Do not treat this list as a working API.

No partial document may be reported as a successful complete render. A diagram
or equation failure must have an explicit profile-level behavior, not silently
lose content. Keep source content authoritative in the host.

## Trust and persistence

Validate source and output within the extension. Independently enforce the
host-approved acceptance policy at the host boundary before persisting or
presenting HTML/SVG/MathML. Rich output needs a reviewed compatible policy;
blindly applying the existing narrow forum sanitizer may remove required rich
markup, while bypassing it would lose the trust boundary.

Preserve BB's forum-specific URL rules, trusted-output types and renderer/search
version semantics through a host adapter. A shared renderer does not own forum
abuse decisions. Public code execution, arbitrary HTML, author scripts and
runtime external resource fetching are not authorized rendering capabilities.

Hosts retain accepted output and required assets so reading does not require a
live render request. Renderer availability controls new render operations, not
ownership of existing content. Revocation, profile upgrades and migrations are
explicit host decisions. Cache isolation/keying must include effective policy
and renderer identity; no cross-tenant private-content cache is implied.

## Lifecycle and compatibility

The planned GOTTH V2 deployment model is one container per extension. This
repository does not implement the host's Compose/lifecycle control path or
grant Docker socket access. Health, admission, routing, unload and cancellation
must use the published V1 control boundary where it applies and an explicitly
specified host-owned lifecycle path where it does not. The generic control
protocol must not be stretched into a renderer or process supervisor.

Start with a BB-compatible profile that leaves Mermaid and math disabled.
Compare source-to-HTML and search projections with BB's existing fixtures.
Rich profiles can then be added without changing that compatibility profile.
Do not extract a renderer by copying BB database, authentication or publication
authority into this extension.

## Costs and unresolved decisions

IPC, process/browser startup, diagram layout, math fonts, output assets and
host validation add costs beyond BB's in-process Markdown renderer. Measure
representative cold/warm requests, concurrency and failure paths before product
integration. No current performance or footprint claim is made.

The implementation specification must settle the separate render protocol,
transport/host compatibility, engine versions/licenses, resource limits, asset
ownership, deterministic IDs, accessibility, syntax delimiters, output profiles
and migration behavior. These are engineering work items, not requests for
blanket new user approval. Adding a module or framework before those decisions
are admitted would only manufacture the appearance of progress.
