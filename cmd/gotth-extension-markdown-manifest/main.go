package main

import (
	"fmt"
	"io"
	"os"

	"github.com/gotthboard/gotth-extension-markdown/internal/markdown"
	extensioncore "github.com/gotthboard/gotth-extensions/pkg/extensions"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: gotth-extension-markdown-manifest <version>")
		os.Exit(2)
	}
	if err := writeManifest(os.Stdout, os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, "manifest generation failed")
		os.Exit(1)
	}
}

func writeManifest(output io.Writer, version string) error {
	manifest, err := markdown.Manifest(version)
	if err != nil {
		return err
	}
	encoded, err := extensioncore.CanonicalManifest(manifest)
	if err != nil {
		return err
	}
	_, err = output.Write(encoded)
	return err
}
