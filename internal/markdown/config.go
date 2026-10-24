package markdown

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	extensioncore "github.com/gotthboard/gotth-extensions/pkg/extensions"
)

const (
	ConfigSchema            = "gotth.extension.markdown.config.v1"
	ExtensionID             = "gotth.render.markdown"
	InterfaceName           = "gotth.markdown.renderer"
	CapabilityRender        = "markdown.render"
	ProfileBBGFM            = "gotth.markdown.bb-gfm.v1"
	ResponseSchema          = "gotth.markdown.render-response.v1"
	MaximumSourceBytes      = 65_536
	MaximumHTMLBytes        = 262_144
	MaximumVisibleTextBytes = 65_536
	MaximumFileBytes        = 65_536
	ProcessMemoryLimitBytes = 128 << 20
)

var BuildVersion = "dev"

const (
	EngineIdentity    = "goldmark-v1.8.5-gfm"
	SanitizerIdentity = "bluemonday-v1.0.27-bb-p2"
)

type Config struct {
	Schema              string `json:"schema"`
	Profile             string `json:"profile"`
	TimeoutMilliseconds int    `json:"timeout_milliseconds"`
	MaxConcurrent       int    `json:"max_concurrent"`
}

type RuntimeConfig struct {
	Config              Config
	Manifest            extensioncore.Manifest
	Grant               extensioncore.Grant
	Host                extensioncore.HostProfile
	Session             extensioncore.Session
	ServiceToken        string
	SocketPath          string
	Version             string
	ConfigurationSHA256 string
}

func (c Config) Timeout() time.Duration {
	return time.Duration(c.TimeoutMilliseconds) * time.Millisecond
}
func LoadRuntimeFromEnv() (RuntimeConfig, error) {
	configPath := os.Getenv("GOTTH_EXTENSION_CONFIG_FILE")
	grantPath := os.Getenv("GOTTH_EXTENSION_GRANT_FILE")
	tokenPath := os.Getenv("GOTTH_EXTENSION_SERVICE_TOKEN_FILE")
	socketPath := os.Getenv("GOTTH_EXTENSION_SOCKET")
	if configPath == "" || grantPath == "" || tokenPath == "" || socketPath == "" || !validVersion(BuildVersion) {
		return RuntimeConfig{}, errors.New("complete extension runtime configuration required")
	}
	configData, err := readPrivateFile(configPath, MaximumFileBytes)
	if err != nil {
		return RuntimeConfig{}, fmt.Errorf("read configuration: %w", err)
	}
	grantData, err := readPrivateFile(grantPath, MaximumFileBytes)
	if err != nil {
		return RuntimeConfig{}, fmt.Errorf("read grant: %w", err)
	}
	tokenData, err := readPrivateFile(tokenPath, 8<<10)
	if err != nil {
		return RuntimeConfig{}, fmt.Errorf("read service token: %w", err)
	}
	config, err := DecodeConfig(configData)
	if err != nil {
		return RuntimeConfig{}, err
	}
	grant, err := DecodeGrant(grantData)
	if err != nil {
		return RuntimeConfig{}, err
	}
	manifest, err := Manifest(BuildVersion)
	if err != nil {
		return RuntimeConfig{}, errors.New("valid artifact manifest required")
	}
	host := MarkdownHostProfile()
	session, err := extensioncore.Negotiate(manifest, grant, host)
	if err != nil || !validRendererSession(session) {
		return RuntimeConfig{}, errors.New("exact Markdown renderer grant required")
	}
	token := string(tokenData)
	clear(tokenData)
	if !validToken(token, 32, 256) {
		return RuntimeConfig{}, errors.New("valid service token required")
	}
	if err := validateSocketPath(socketPath); err != nil {
		return RuntimeConfig{}, err
	}
	digest := sha256.Sum256(configData)
	return RuntimeConfig{Config: config, Manifest: manifest, Grant: grant, Host: host, Session: session, ServiceToken: token, SocketPath: socketPath, Version: BuildVersion, ConfigurationSHA256: hex.EncodeToString(digest[:])}, nil
}

func DecodeConfig(data []byte) (Config, error) {
	fields, err := decodeUniqueObject(data)
	if err != nil || len(fields) != 4 {
		return Config{}, errors.New("invalid extension configuration")
	}
	var c Config
	if decodeField(fields, "schema", &c.Schema) != nil || c.Schema != ConfigSchema {
		return Config{}, errors.New("invalid configuration schema")
	}
	if decodeField(fields, "profile", &c.Profile) != nil || c.Profile != ProfileBBGFM {
		return Config{}, errors.New("unsupported rendering profile")
	}
	if decodeField(fields, "timeout_milliseconds", &c.TimeoutMilliseconds) != nil || c.TimeoutMilliseconds < 50 || c.TimeoutMilliseconds > 2_000 {
		return Config{}, errors.New("timeout must be between 50 and 2000 milliseconds")
	}
	if decodeField(fields, "max_concurrent", &c.MaxConcurrent) != nil || c.MaxConcurrent < 1 || c.MaxConcurrent > 8 {
		return Config{}, errors.New("max_concurrent must be between 1 and 8")
	}
	canonical, err := json.Marshal(c)
	if err != nil || !bytes.Equal(data, canonical) {
		return Config{}, errors.New("canonical extension configuration required")
	}
	return c, nil
}

func DecodeGrant(data []byte) (extensioncore.Grant, error) {
	grant, err := extensioncore.ParseGrant(data)
	if err != nil {
		return extensioncore.Grant{}, errors.New("invalid extension grant")
	}
	canonical, err := extensioncore.CanonicalGrant(grant)
	if err != nil || !bytes.Equal(data, canonical) {
		return extensioncore.Grant{}, errors.New("canonical extension grant required")
	}
	return grant, nil
}

func MarkdownHostProfile() extensioncore.HostProfile {
	return extensioncore.HostProfile{
		Protocols:  []extensioncore.VersionRange{{Name: extensioncore.ControlName, Major: 1, MinMinor: 0, MaxMinor: 0}},
		Interfaces: []extensioncore.VersionRange{{Name: InterfaceName, Major: 1, MinMinor: 0, MaxMinor: 0}},
	}
}

func validRendererSession(session extensioncore.Session) bool {
	return session.ExtensionID == ExtensionID &&
		session.Control == (extensioncore.ProtocolVersion{Name: extensioncore.ControlName, Major: 1, Minor: 0}) &&
		len(session.Interfaces) == 1 && session.Interfaces[0] == (extensioncore.InterfaceGrant{Name: InterfaceName, Major: 1, Minor: 0}) &&
		len(session.Capabilities) == 1 && session.Capabilities[0] == CapabilityRender &&
		len(session.Secrets) == 0
}

func decodeUniqueObject(data []byte) (map[string]json.RawMessage, error) {
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	t, err := d.Token()
	if err != nil || t != json.Delim('{') {
		return nil, errors.New("JSON object required")
	}
	fields := map[string]json.RawMessage{}
	for d.More() {
		t, err := d.Token()
		name, ok := t.(string)
		if err != nil || !ok {
			return nil, errors.New("invalid JSON object")
		}
		if _, ok := fields[name]; ok {
			return nil, errors.New("duplicate JSON field")
		}
		var raw json.RawMessage
		if err := d.Decode(&raw); err != nil {
			return nil, err
		}
		fields[name] = raw
	}
	if t, err = d.Token(); err != nil || t != json.Delim('}') {
		return nil, errors.New("invalid JSON object")
	}
	if err := d.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return nil, errors.New("trailing JSON input")
	}
	return fields, nil
}
func decodeField(fields map[string]json.RawMessage, name string, target any) error {
	raw, ok := fields[name]
	if !ok {
		return errors.New("missing field")
	}
	return json.Unmarshal(raw, target)
}

func readPrivateFile(path string, limit int64) ([]byte, error) {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return nil, errors.New("canonical absolute private file path required")
	}
	if err := validatePrivateDir(filepath.Dir(path)); err != nil {
		return nil, errors.New("private file directory required")
	}
	info, err := os.Lstat(path)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 || !ownedByEUID(info) {
		return nil, errors.New("private owned regular file required")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	opened, err := f.Stat()
	if err != nil || !os.SameFile(info, opened) || !opened.Mode().IsRegular() || opened.Mode().Perm()&0o077 != 0 || !ownedByEUID(opened) {
		return nil, errors.New("private regular file changed during open")
	}
	data, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil || int64(len(data)) > limit {
		return nil, errors.New("bounded private file required")
	}
	return data, nil
}
func ownedByEUID(info os.FileInfo) bool {
	st, ok := info.Sys().(*syscall.Stat_t)
	return ok && st.Uid == uint32(os.Geteuid())
}
func validatePrivateDir(path string) error {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return errors.New("canonical absolute private directory path required")
	}
	info, err := os.Lstat(path)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() || info.Mode().Perm()&0o077 != 0 || !ownedByEUID(info) {
		return errors.New("private owned directory required")
	}
	d, err := os.Open(path)
	if err != nil {
		return errors.New("private owned directory required")
	}
	defer d.Close()
	opened, err := d.Stat()
	if err != nil || !os.SameFile(info, opened) || !opened.IsDir() || opened.Mode().Perm()&0o077 != 0 || !ownedByEUID(opened) {
		return errors.New("private directory changed during open")
	}
	return nil
}
func validateSocketPath(path string) error {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path || len(path) > 100 {
		return errors.New("bounded absolute Unix socket path required")
	}
	if err := validatePrivateDir(filepath.Dir(path)); err != nil {
		return errors.New("private Unix socket directory required")
	}
	if _, err := os.Lstat(path); err == nil || !errors.Is(err, os.ErrNotExist) {
		return errors.New("Unix socket path must not exist")
	}
	return nil
}
func validVersion(v string) bool {
	if v == "dev" || v == "1.0.0" {
		return true
	}
	for _, p := range []string{"1.0.0-alpha.", "1.0.0-beta."} {
		if strings.HasPrefix(v, p) {
			n := v[len(p):]
			return n != "" && n[0] != '0' && strings.Trim(n, "0123456789") == ""
		}
	}
	return false
}
func validToken(v string, min, max int) bool {
	if len(v) < min || len(v) > max {
		return false
	}
	for i := range len(v) {
		if v[i] < 0x21 || v[i] > 0x7e {
			return false
		}
	}
	return true
}
func validHex(v string, n int) bool {
	if len(v) != n {
		return false
	}
	for _, r := range v {
		if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f') {
			return false
		}
	}
	return true
}
func validUUID(v string) bool {
	return len(v) == 36 && v[8] == '-' && v[13] == '-' && v[18] == '-' && v[23] == '-' && validHex(strings.ReplaceAll(v, "-", ""), 32)
}
