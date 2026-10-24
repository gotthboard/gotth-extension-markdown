# Coverage map

| Requirement/surface | Mechanism | Tests/checks | Open gate |
|---|---|---|---|
| MD-001/003 profile boundary | one manifest/interface/profile; rich syntax remains text/code | `TestV1DoesNotEnableFutureProfiles`; source-boundary scan | future profiles stay separate |
| MD-002 BB compatibility | exact Goldmark 1.8.5 GFM and BB p2 sanitizer | exact HTML fixture and hostile cases | host-owned BB fixture/migration run is outside this repository |
| MD-004 deterministic identity | canonical identity includes engine/profile/sanitizer/config/assets/limits | repeat and identity tests; reproducible archive | final source-bound evidence |
| MD-005/007 host authority | host-called renderer, no callbacks/persistence/trusted type | protocol review; no-egress scan | consuming host acceptance remains separate |
| MD-006 bounds/failure | source/output/text/message/time/memory/concurrency bounds | unit, expansion, synchronized server-observed cancellation, process, race/repeat/fuzz evidence | two fresh independent reviews |
| MD-008 availability | host-retained source/output; no fallback contract | documentation review | consuming host implementation |
| MD-009 accessibility | semantic sanitized HTML and visible-text projection | exact headings/lists/tasks/table/text fixture | richer diagrams/equations not V1 |
| MD-010/011 foundation | exact v1.0.0 modules and SDK control imports | source SHA, descriptor SHA/RPC, no-local-copy checks | none in source candidate |
| MD-012 local lifecycle | private files/socket, token/deadline/correlation, graceful stop | config/service/compiled-process and development race evidence | two fresh independent reviews |
| MD-013 supply chain | checksums, retained licenses/notices, SPDX, pinned tools/environment, manifest-bound gates, offline cache-seeded build, deterministic validated archive | dependency/policy/artifact evidence; advisory scan completed with no findings | clean-commit artifact and two fresh independent reviews |
