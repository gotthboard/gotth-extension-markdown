package main

import (
	"bytes"
	"testing"

	"github.com/gotthboard/gotth-extension-markdown/internal/markdown"
	extensioncore "github.com/gotthboard/gotth-extensions/pkg/extensions"
)

func TestWriteManifestIsByteExactCanonicalDocument(t *testing.T) {
	manifest, err := markdown.Manifest("1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	want, err := extensioncore.CanonicalManifest(manifest)
	if err != nil {
		t.Fatal(err)
	}
	var got bytes.Buffer
	if err := writeManifest(&got, "1.0.0"); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got.Bytes(), want) {
		t.Fatalf("manifest bytes differ: got=%q want=%q", got.Bytes(), want)
	}
	if bytes.HasSuffix(got.Bytes(), []byte("\n")) {
		t.Fatal("canonical manifest has trailing newline")
	}
}
