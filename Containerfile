FROM scratch
COPY gotth-extension-markdown /gotth-extension-markdown
COPY LICENSE /usr/share/licenses/gotth-extension-markdown/LICENSE
COPY THIRD_PARTY_NOTICES.md /usr/share/licenses/gotth-extension-markdown/THIRD_PARTY_NOTICES.md
COPY THIRD_PARTY_LICENSES /usr/share/licenses/gotth-extension-markdown/THIRD_PARTY_LICENSES
COPY SBOM.spdx.json /usr/share/doc/gotth-extension-markdown/SBOM.spdx.json
USER 65532:65532
ENTRYPOINT ["/gotth-extension-markdown"]
