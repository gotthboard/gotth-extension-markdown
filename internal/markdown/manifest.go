package markdown

import extensioncore "github.com/gotthboard/gotth-extensions/pkg/extensions"

func Manifest(version string) (extensioncore.Manifest, error) {
	if !validVersion(version) {
		return extensioncore.Manifest{}, extensioncore.ErrInvalidInput
	}
	manifestVersion := version
	if version == "dev" {
		manifestVersion = "0.0.0-dev"
	}
	manifest := extensioncore.Manifest{
		Schema:       extensioncore.ManifestSchema,
		ID:           ExtensionID,
		Name:         "GOTTH Markdown Renderer",
		Version:      manifestVersion,
		Protocols:    []extensioncore.VersionRange{{Name: extensioncore.ControlName, Major: 1, MinMinor: 0, MaxMinor: 0}},
		Interfaces:   []extensioncore.VersionRange{{Name: InterfaceName, Major: 1, MinMinor: 0, MaxMinor: 0}},
		Capabilities: []string{CapabilityRender},
		Secrets:      []extensioncore.SecretRequirement{},
	}
	if err := extensioncore.ValidateManifest(manifest); err != nil {
		return extensioncore.Manifest{}, err
	}
	return manifest, nil
}

func ManifestDigest(version string) (string, error) {
	manifest, err := Manifest(version)
	if err != nil {
		return "", err
	}
	return extensioncore.ManifestDigest(manifest)
}
