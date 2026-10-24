#!/bin/sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd -P)
. "$root/scripts/pinned-toolchain.sh"
. "$root/scripts/rg-policy-lib.sh"
cd "$root"
production="cmd internal"
rg_reject_matches /usr/bin/rg 'subprocess or plugin mechanism is forbidden' -n '"os/exec"|"plugin"|exec\.(Command|CommandContext)\b|os\.StartProcess\b|syscall\.(Exec|ForkExec|StartProcess)\b|unix\.(Exec|ForkExec|StartProcess)\b' $production --glob '*.go' --glob '!**/*_test.go'
rg_reject_matches /usr/bin/rg 'outbound network client mechanism is forbidden' -n '"net/http"|"net/smtp"|"net/rpc"|http\.(Get|Head|Post|PostForm|Do|Client|DefaultClient|Transport|NewRequest|NewRequestWithContext)\b|net\.(Dial|DialTimeout|DialTCP|DialUDP|DialUnix|Dialer|Resolver|ResolveTCPAddr|ResolveUDPAddr)\b|tls\.(Dial|DialWithDialer)\b|grpc\.NewClient\b|ProxyFromEnvironment|Getenv\("(http|https|all|no)_proxy"' $production --glob '*.go' --glob '!**/*_test.go'
rg_reject_matches /usr/bin/rg 'TCP, UDP, IP, or HTTP listener mechanism is forbidden' --pcre2 -n 'net\.(ListenTCP|ListenUDP|ListenIP|ListenPacket|ListenConfig)\b|net\.Listen\((?!"unix")|tls\.Listen\b|http\.(ListenAndServe|ListenAndServeTLS|Server)\b' $production --glob '*.go' --glob '!**/*_test.go'
rg_reject_matches /usr/bin/rg 'future rich renderer leaked into V1 production surface' -n 'mermaid|katex|mathjax' cmd internal proto --glob '*.go' --glob '*.proto' --glob '!**/*_test.go'
[ "$(/usr/bin/rg -n 'net\.Listen\("unix", runtimeConfig\.SocketPath\)' cmd/gotth-extension-markdown/main.go | wc -l)" -eq 1 ] || { echo 'the one admitted Unix listener changed' >&2; exit 1; }
[ "$(find proto/gotth/markdown/renderer/v1 -maxdepth 1 -type f | wc -l)" -eq 3 ]
printf '%s\n' 'production source has no subprocess, plugin, outbound-client, or TCP/UDP/IP/HTTP listener mechanism; exactly one Unix listener is admitted'
