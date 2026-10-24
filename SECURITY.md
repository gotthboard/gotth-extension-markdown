# Security policy

## Reporting

Report vulnerabilities through
[GitHub private vulnerability reporting](https://github.com/gotthboard/gotth-extension-markdown/security/advisories/new).
Do not include live credentials or private content, and do not use
[public issues](https://github.com/gotthboard/gotth-extension-markdown/issues),
pull requests, discussions, or Forgejo tickets for security reports. The
[canonical source](https://git.dannyhunn.com/gotthboard/gotth-extension-markdown)
and [public mirror](https://github.com/gotthboard/gotth-extension-markdown) are
directly linked here so reports are not routed by guesswork.

## Supported status

There is no published supported release yet. The repository contains a v1.0.0
source candidate; a source branch or locally built archive is not a release.

## Security boundary

The host owns identity, authorization, source, persistence, publication,
tenant/cache isolation, and final output policy. The extension accepts only an
authenticated deadline-bearing local Unix-socket call, renders one bounded
profile in process, and returns untrusted sanitized HTML plus visible text. It
has no outbound client, subprocess, browser, database, product callback,
container-engine access, or runtime asset fetch.

Deploy with no network, read-only root, private owner-only runtime files/socket,
non-root UID 65532, and hard resource limits. Rotate the local service token and
replace the private runtime directory on each process session. Treat a profile,
engine, sanitizer, configuration, asset, or host-policy identity change as an
explicit migration, never a cache hit.
