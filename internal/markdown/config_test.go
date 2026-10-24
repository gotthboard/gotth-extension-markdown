package markdown

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	extensioncore "github.com/gotthboard/gotth-extensions/pkg/extensions"
)

func TestDecodeConfigAndGrantStrict(t *testing.T) {
	validConfig := []byte("{\"schema\":\"gotth.extension.markdown.config.v1\",\"profile\":\"gotth.markdown.bb-gfm.v1\",\"timeout_milliseconds\":500,\"max_concurrent\":4}")
	config, err := DecodeConfig(validConfig)
	if err != nil || config.TimeoutMilliseconds != 500 || config.MaxConcurrent != 4 {
		t.Fatalf("config=%+v err=%v", config, err)
	}
	invalid := [][]byte{
		[]byte("{}"),
		[]byte("{\"schema\":\"gotth.extension.markdown.config.v1\",\"schema\":\"gotth.extension.markdown.config.v1\",\"profile\":\"gotth.markdown.bb-gfm.v1\",\"timeout_milliseconds\":500,\"max_concurrent\":4}"),
		append(append([]byte{}, validConfig...), []byte("{}")...),
		[]byte("{\"schema\":\"gotth.extension.markdown.config.v1\",\"profile\":\"gotth.markdown.rich.v1\",\"timeout_milliseconds\":500,\"max_concurrent\":4}"),
		append(append([]byte{}, validConfig...), '\n'),
		[]byte("{ \"schema\":\"gotth.extension.markdown.config.v1\",\"profile\":\"gotth.markdown.bb-gfm.v1\",\"timeout_milliseconds\":500,\"max_concurrent\":4}"),
		[]byte("{\"profile\":\"gotth.markdown.bb-gfm.v1\",\"schema\":\"gotth.extension.markdown.config.v1\",\"timeout_milliseconds\":500,\"max_concurrent\":4}"),
		[]byte("{\"schema\":\"gotth.extension.markdown.config.v\\u0031\",\"profile\":\"gotth.markdown.bb-gfm.v1\",\"timeout_milliseconds\":500,\"max_concurrent\":4}"),
	}
	for i, raw := range invalid {
		if _, err := DecodeConfig(raw); err == nil {
			t.Fatalf("invalid config %d accepted", i)
		}
	}
	runtimeConfig := testRuntime()
	canonical, err := extensioncore.CanonicalGrant(runtimeConfig.Grant)
	if err != nil {
		t.Fatal(err)
	}
	grant, err := DecodeGrant(canonical)
	if err != nil || grant.InstanceID != runtimeConfig.Grant.InstanceID {
		t.Fatalf("grant=%+v err=%v", grant, err)
	}
	for _, raw := range [][]byte{
		append(append([]byte{}, canonical...), '\n'),
		append(append([]byte{}, canonical...), []byte("{}")...),
		[]byte(strings.Replace(string(canonical), "{", "{\n  ", 1)),
	} {
		if _, err := DecodeGrant(raw); err == nil {
			t.Fatalf("noncanonical grant accepted: %q", raw)
		}
	}
}

func FuzzDecodeConfigRejectsUntrustedShapes(f *testing.F) {
	f.Add([]byte(`{"schema":"gotth.extension.markdown.config.v1","profile":"gotth.markdown.bb-gfm.v1","timeout_milliseconds":500,"max_concurrent":4}`))
	f.Add([]byte(`{"schema":"x","source":"private"}`))
	f.Fuzz(func(t *testing.T, data []byte) {
		config, err := DecodeConfig(data)
		if err == nil && (config.Schema != ConfigSchema || config.Profile != ProfileBBGFM || config.TimeoutMilliseconds < 50 || config.TimeoutMilliseconds > 2000 || config.MaxConcurrent < 1 || config.MaxConcurrent > 8) {
			t.Fatalf("invalid config accepted: %+v", config)
		}
	})
}

func TestLoadRuntimeRequiresPrivateFilesAndExactGrant(t *testing.T) {
	root := t.TempDir()
	if err := os.Chmod(root, 0o700); err != nil {
		t.Fatal(err)
	}
	runtimeConfig := testRuntime()
	grant, err := extensioncore.CanonicalGrant(runtimeConfig.Grant)
	if err != nil {
		t.Fatal(err)
	}
	config := []byte("{\"schema\":\"gotth.extension.markdown.config.v1\",\"profile\":\"gotth.markdown.bb-gfm.v1\",\"timeout_milliseconds\":500,\"max_concurrent\":4}")
	files := map[string][]byte{"config": config, "grant": grant, "token": []byte(testToken)}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(root, name), body, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	socket := filepath.Join(root, "renderer.sock")
	t.Setenv("GOTTH_EXTENSION_CONFIG_FILE", filepath.Join(root, "config"))
	t.Setenv("GOTTH_EXTENSION_GRANT_FILE", filepath.Join(root, "grant"))
	t.Setenv("GOTTH_EXTENSION_SERVICE_TOKEN_FILE", filepath.Join(root, "token"))
	t.Setenv("GOTTH_EXTENSION_SOCKET", socket)
	loaded, err := LoadRuntimeFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.ConfigurationSHA256 == "" || loaded.SocketPath != socket || loaded.Session.Fingerprint == "" || !validRendererSession(loaded.Session) {
		t.Fatalf("runtime=%+v", loaded)
	}
	if err := os.Chmod(filepath.Join(root, "token"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadRuntimeFromEnv(); err == nil || strings.Contains(err.Error(), testToken) {
		t.Fatalf("insecure token error=%v", err)
	}
	if err := os.Chmod(filepath.Join(root, "token"), 0o600); err != nil {
		t.Fatal(err)
	}
	configLink := filepath.Join(root, "config-link")
	if err := os.Symlink(filepath.Join(root, "config"), configLink); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOTTH_EXTENSION_CONFIG_FILE", configLink)
	if _, err := LoadRuntimeFromEnv(); err == nil {
		t.Fatal("symlinked configuration accepted")
	}
	t.Setenv("GOTTH_EXTENSION_CONFIG_FILE", filepath.Join(root, "config"))
	if err := os.WriteFile(socket, []byte("occupied"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadRuntimeFromEnv(); err == nil {
		t.Fatal("preexisting socket path accepted")
	}
}

func TestServiceTokenGrammarIsPrintableASCII(t *testing.T) {
	if !validToken(strings.Repeat("!", 32), 32, 256) || !validToken(strings.Repeat("~", 32), 32, 256) {
		t.Fatal("printable ASCII boundary rejected")
	}
	for _, token := range []string{strings.Repeat("a", 31), strings.Repeat("a", 257), strings.Repeat("a", 31) + " ", strings.Repeat("a", 31) + "\x7f", strings.Repeat("a", 31) + "é"} {
		if validToken(token, 32, 256) {
			t.Fatalf("invalid token accepted: %q", token)
		}
	}
}
