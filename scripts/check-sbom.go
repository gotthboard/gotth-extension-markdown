//go:build ignore

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

type document struct {
	SPDXVersion       string         `json:"spdxVersion"`
	DocumentDescribes []string       `json:"documentDescribes"`
	Packages          []spdxPackage  `json:"packages"`
	Relationships     []relationship `json:"relationships"`
}
type spdxPackage struct {
	Name             string        `json:"name"`
	SPDXID           string        `json:"SPDXID"`
	Version          string        `json:"versionInfo"`
	LicenseConcluded string        `json:"licenseConcluded"`
	LicenseDeclared  string        `json:"licenseDeclared"`
	ExternalRefs     []externalRef `json:"externalRefs"`
}
type externalRef struct {
	Category string `json:"referenceCategory"`
	Type     string `json:"referenceType"`
	Locator  string `json:"referenceLocator"`
}
type relationship struct {
	From string `json:"spdxElementId"`
	Type string `json:"relationshipType"`
	To   string `json:"relatedSpdxElement"`
}
type expectedPackage struct{ version, license, purl string }

func fail(format string, values ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", values...)
	os.Exit(1)
}

func main() {
	if len(os.Args) != 2 {
		fail("usage: check-sbom.go <SBOM.spdx.json>")
	}
	modules := map[string]string{}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 1 || len(fields) > 2 {
			fail("invalid linked module input: %q", scanner.Text())
		}
		version := ""
		if len(fields) == 2 {
			version = fields[1]
		}
		if _, duplicate := modules[fields[0]]; duplicate {
			fail("duplicate linked module %s", fields[0])
		}
		modules[fields[0]] = version
	}
	if err := scanner.Err(); err != nil {
		fail("read linked modules: %v", err)
	}
	licenses := map[string]string{
		"github.com/gotthboard/gotth-extension-markdown": "MIT",
		"github.com/aymerick/douceur":                    "MIT", "github.com/gorilla/css": "BSD-3-Clause",
		"github.com/gotthboard/gotth-extensions": "MIT", "github.com/gotthboard/gotth-sdk": "MIT",
		"github.com/microcosm-cc/bluemonday": "BSD-3-Clause", "github.com/yuin/goldmark": "MIT",
		"golang.org/x/net": "BSD-3-Clause", "golang.org/x/sys": "BSD-3-Clause", "golang.org/x/text": "BSD-3-Clause",
		"google.golang.org/genproto/googleapis/rpc": "Apache-2.0", "google.golang.org/grpc": "Apache-2.0",
		"google.golang.org/protobuf": "BSD-3-Clause",
	}
	expected := map[string]expectedPackage{}
	for name, version := range modules {
		license, ok := licenses[name]
		if !ok {
			fail("no admitted SPDX license for linked module %s", name)
		}
		spdxName := name
		spdxVersion := version
		purlVersion := version
		if name == "github.com/gotthboard/gotth-extension-markdown" {
			spdxName, spdxVersion, purlVersion = "gotth-extension-markdown", "1.0.0", "v1.0.0"
		}
		expected[spdxName] = expectedPackage{spdxVersion, license, "pkg:golang/" + name + "@" + purlVersion}
	}
	expected["Go standard library"] = expectedPackage{"go1.27.3-X:nodwarf5", "BSD-3-Clause", "pkg:generic/go-stdlib@1.27.3-X:nodwarf5"}

	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fail("read SBOM: %v", err)
	}
	var doc document
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	if err := decoder.Decode(&doc); err != nil {
		fail("parse SBOM: %v", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		fail("parse SBOM: trailing JSON input")
	}
	if doc.SPDXVersion != "SPDX-2.3" || len(doc.DocumentDescribes) != 1 || doc.DocumentDescribes[0] != "SPDXRef-Package-root" {
		fail("invalid SPDX document identity")
	}
	if len(doc.Packages) != len(expected) {
		fail("SBOM package count=%d want=%d", len(doc.Packages), len(expected))
	}
	ids := map[string]string{}
	for _, pkg := range doc.Packages {
		want, ok := expected[pkg.Name]
		if !ok {
			fail("unexpected SBOM package %s", pkg.Name)
		}
		if _, duplicate := ids[pkg.SPDXID]; duplicate || pkg.SPDXID == "" {
			fail("duplicate or empty SPDX package ID %s", pkg.SPDXID)
		}
		ids[pkg.SPDXID] = pkg.Name
		if pkg.Version != want.version || pkg.LicenseConcluded != want.license || pkg.LicenseDeclared != want.license {
			fail("SBOM metadata mismatch for %s: version=%q concluded=%q declared=%q", pkg.Name, pkg.Version, pkg.LicenseConcluded, pkg.LicenseDeclared)
		}
		if len(pkg.ExternalRefs) != 1 || pkg.ExternalRefs[0] != (externalRef{Category: "PACKAGE-MANAGER", Type: "purl", Locator: want.purl}) {
			fail("SBOM purl mismatch for %s", pkg.Name)
		}
		delete(expected, pkg.Name)
	}
	if len(expected) != 0 {
		missing := make([]string, 0, len(expected))
		for name := range expected {
			missing = append(missing, name)
		}
		sort.Strings(missing)
		fail("SBOM missing packages: %s", strings.Join(missing, ", "))
	}
	rootID := "SPDXRef-Package-root"
	wantRelationships := map[string]struct{}{}
	for id := range ids {
		if id == rootID {
			continue
		}
		wantRelationships[rootID+"\x00STATIC_LINK\x00"+id] = struct{}{}
	}
	if len(doc.Relationships) != len(wantRelationships) {
		fail("SBOM relationship count=%d want=%d", len(doc.Relationships), len(wantRelationships))
	}
	for _, rel := range doc.Relationships {
		key := rel.From + "\x00" + rel.Type + "\x00" + rel.To
		if _, ok := wantRelationships[key]; !ok {
			fail("unexpected SBOM relationship %s %s %s", rel.From, rel.Type, rel.To)
		}
		delete(wantRelationships, key)
	}
	if len(wantRelationships) != 0 {
		fail("SBOM relationships missing")
	}
	fmt.Println("exact linked dependency/version/license/purl/relationship parity verified")
}
