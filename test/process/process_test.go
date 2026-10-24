package process_test

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/gotthboard/gotth-extension-markdown/internal/markdown"
	rendererv1 "github.com/gotthboard/gotth-extension-markdown/proto/gotth/markdown/renderer/v1"
	extensioncore "github.com/gotthboard/gotth-extensions/pkg/extensions"
	extensionsv1 "github.com/gotthboard/gotth-sdk/gen/go/gotth/extensions/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type rawProtoCodec struct{ payload []byte }

func (c rawProtoCodec) Name() string                { return "proto" }
func (c rawProtoCodec) Marshal(any) ([]byte, error) { return append([]byte(nil), c.payload...), nil }
func (c rawProtoCodec) Unmarshal(data []byte, value any) error {
	message, ok := value.(proto.Message)
	if !ok {
		return fmt.Errorf("raw protobuf response type %T is not a protobuf message", value)
	}
	return proto.Unmarshal(data, message)
}

const token = "12345678901234567890123456789012"

func TestAuthenticatedUnixProcess(t *testing.T) {
	binary := os.Getenv("GOTTH_MARKDOWN_BINARY")
	if binary == "" {
		t.Skip("GOTTH_MARKDOWN_BINARY is required")
	}
	root := t.TempDir()
	if err := os.Chmod(root, 0o700); err != nil {
		t.Fatal(err)
	}
	manifest, err := markdown.Manifest("dev")
	if err != nil {
		t.Fatal(err)
	}
	manifestDigest, err := extensioncore.ManifestDigest(manifest)
	if err != nil {
		t.Fatal(err)
	}
	grant := extensioncore.Grant{Schema: extensioncore.GrantSchema, InstanceID: "11111111-1111-1111-1111-111111111111", ExtensionID: markdown.ExtensionID, ManifestDigest: manifestDigest, Capabilities: []string{markdown.CapabilityRender}, Interfaces: []extensioncore.InterfaceGrant{{Name: markdown.InterfaceName, Major: 1, Minor: 0}}, Secrets: []string{}}
	grantData, err := extensioncore.CanonicalGrant(grant)
	if err != nil {
		t.Fatal(err)
	}
	session, err := extensioncore.Negotiate(manifest, grant, markdown.MarkdownHostProfile())
	if err != nil {
		t.Fatal(err)
	}
	config := []byte("{\"schema\":\"gotth.extension.markdown.config.v1\",\"profile\":\"gotth.markdown.bb-gfm.v1\",\"timeout_milliseconds\":500,\"max_concurrent\":2}")
	for name, body := range map[string][]byte{"config": config, "grant": grantData, "token": []byte(token)} {
		if err := os.WriteFile(filepath.Join(root, name), body, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	socket := filepath.Join(root, "renderer.sock")
	cmd := exec.Command(binary)
	cmd.Env = []string{"GOTTH_EXTENSION_CONFIG_FILE=" + filepath.Join(root, "config"), "GOTTH_EXTENSION_GRANT_FILE=" + filepath.Join(root, "grant"), "GOTTH_EXTENSION_SERVICE_TOKEN_FILE=" + filepath.Join(root, "token"), "GOTTH_EXTENSION_SOCKET=" + socket}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	t.Cleanup(func() {
		_ = cmd.Process.Signal(syscall.SIGTERM)
		select {
		case err := <-done:
			if err != nil {
				t.Errorf("process exit: %v", err)
			}
		case <-time.After(3 * time.Second):
			_ = cmd.Process.Kill()
			<-done
			t.Error("process did not stop gracefully")
		}
	})
	deadline := time.Now().Add(3 * time.Second)
	for {
		if _, err := os.Lstat(socket); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("renderer socket did not appear")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if info, err := os.Lstat(socket); err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("socket permissions: %v %v", info, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	conn, err := grpc.NewClient("passthrough:///renderer", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "unix", socket)
	}))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	control := extensionsv1.NewExtensionControlClient(conn)
	healthRequest := &extensionsv1.HealthRequest{CorrelationId: "22222222-2222-2222-2222-222222222222", SessionSha256: session.Fingerprint}
	client := rendererv1.NewMarkdownRendererClient(conn)
	for name, invoke := range map[string]func(context.Context) error{
		"handshake": func(callContext context.Context) error {
			_, err := control.Handshake(callContext, &extensionsv1.HandshakeRequest{})
			return err
		},
		"health": func(callContext context.Context) error {
			_, err := control.Health(callContext, &extensionsv1.HealthRequest{})
			return err
		},
		"render": func(callContext context.Context) error {
			_, err := client.Render(callContext, &rendererv1.RenderRequest{})
			return err
		},
	} {
		if err := invoke(ctx); status.Code(err) != codes.Unauthenticated {
			t.Fatalf("malformed unauthenticated %s code=%v err=%v", name, status.Code(err), err)
		}
	}
	if _, err := control.Health(ctx, healthRequest); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("unauthenticated call: %v", err)
	}
	authed := metadata.AppendToOutgoingContext(ctx, "x-gotth-service-token", token)
	short := metadata.AppendToOutgoingContext(ctx, "x-gotth-service-token", "too-short")
	if _, err := control.Health(short, healthRequest); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("unequal-length token call: %v", err)
	}
	health, err := control.Health(authed, healthRequest)
	if err != nil || health.GetState() != extensionsv1.HealthState_HEALTH_STATE_READY {
		t.Fatalf("health=%+v err=%v", health, err)
	}
	challenge := []byte("01234567890123456789012345678901")
	handshake, err := control.Handshake(authed, &extensionsv1.HandshakeRequest{CorrelationId: "44444444-4444-4444-4444-444444444444", Challenge: challenge, ExpectedExtensionId: session.ExtensionID, ExpectedManifestSha256: session.ManifestDigest, ExpectedGrantSha256: session.GrantDigest, ExpectedSessionSha256: session.Fingerprint})
	if err != nil || string(handshake.GetChallenge()) != string(challenge) || len(handshake.GetInterfaces()) != 1 || handshake.GetInterfaces()[0].GetName() != markdown.InterfaceName {
		t.Fatalf("handshake=%+v err=%v", handshake, err)
	}
	response, err := client.Render(authed, &rendererv1.RenderRequest{CorrelationId: "33333333-3333-3333-3333-333333333333", Profile: markdown.ProfileBBGFM, Source: "# Safe\n\n<script>secret</script>"})
	if err != nil {
		t.Fatal(err)
	}
	if !response.GetComplete() || response.GetExtensionVersion() != "dev" || response.GetVisibleText() != "Safe" || response.GetHtml() != "<h1>Safe</h1>\n\n" {
		t.Fatalf("render=%+v", response)
	}
	semantic, err := client.Render(authed, &rendererv1.RenderRequest{CorrelationId: "77777777-7777-4777-8777-777777777777", Profile: markdown.ProfileBBGFM, Source: "---"})
	if err != nil || !semantic.GetComplete() || semantic.GetHtml() != "<hr>\n" || semantic.GetVisibleText() != "" {
		t.Fatalf("semantic render=%+v err=%v", semantic, err)
	}
	_, err = client.Render(authed, &rendererv1.RenderRequest{CorrelationId: "55555555-5555-5555-5555-555555555555", Profile: markdown.ProfileBBGFM, Source: strings.Repeat("x", 100_000)})
	if status.Code(err) != codes.ResourceExhausted {
		t.Fatalf("oversized transport code=%v err=%v", status.Code(err), err)
	}
	var malformedResponse rendererv1.RenderResponse
	err = conn.Invoke(authed, "/gotth.markdown.renderer.v1.MarkdownRenderer/Render", struct{}{}, &malformedResponse, grpc.ForceCodec(rawProtoCodec{payload: []byte{0x1a, 0x02, 0xff, 0xff}}))
	if status.Code(err) != codes.Internal {
		t.Fatalf("malformed protobuf decode code=%v err=%v", status.Code(err), err)
	}
	canceled, stop := context.WithCancel(authed)
	stop()
	_, err = client.Render(canceled, &rendererv1.RenderRequest{CorrelationId: "66666666-6666-6666-6666-666666666666", Profile: markdown.ProfileBBGFM, Source: "body"})
	if status.Code(err) != codes.Canceled {
		t.Fatalf("canceled code=%v err=%v", status.Code(err), err)
	}
}

func TestProcessRejectsNonASCIITokenBeforeListen(t *testing.T) {
	binary := os.Getenv("GOTTH_MARKDOWN_BINARY")
	if binary == "" {
		t.Skip("GOTTH_MARKDOWN_BINARY is required")
	}
	root := t.TempDir()
	if err := os.Chmod(root, 0o700); err != nil {
		t.Fatal(err)
	}
	manifest, err := markdown.Manifest("dev")
	if err != nil {
		t.Fatal(err)
	}
	manifestDigest, err := extensioncore.ManifestDigest(manifest)
	if err != nil {
		t.Fatal(err)
	}
	grant := extensioncore.Grant{Schema: extensioncore.GrantSchema, InstanceID: "11111111-1111-1111-1111-111111111111", ExtensionID: markdown.ExtensionID, ManifestDigest: manifestDigest, Capabilities: []string{markdown.CapabilityRender}, Interfaces: []extensioncore.InterfaceGrant{{Name: markdown.InterfaceName, Major: 1, Minor: 0}}, Secrets: []string{}}
	grantData, err := extensioncore.CanonicalGrant(grant)
	if err != nil {
		t.Fatal(err)
	}
	config := []byte("{\"schema\":\"gotth.extension.markdown.config.v1\",\"profile\":\"gotth.markdown.bb-gfm.v1\",\"timeout_milliseconds\":500,\"max_concurrent\":2}")
	for name, body := range map[string][]byte{"config": config, "grant": grantData, "token": []byte(strings.Repeat("a", 31) + "é")} {
		if err := os.WriteFile(filepath.Join(root, name), body, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	socket := filepath.Join(root, "renderer.sock")
	cmd := exec.Command(binary)
	cmd.Env = []string{"GOTTH_EXTENSION_CONFIG_FILE=" + filepath.Join(root, "config"), "GOTTH_EXTENSION_GRANT_FILE=" + filepath.Join(root, "grant"), "GOTTH_EXTENSION_SERVICE_TOKEN_FILE=" + filepath.Join(root, "token"), "GOTTH_EXTENSION_SOCKET=" + socket}
	output, err := cmd.CombinedOutput()
	if err == nil || strings.Contains(string(output), strings.Repeat("a", 31)) {
		t.Fatalf("non-ASCII token startup err=%v output=%q", err, output)
	}
	if _, err := os.Lstat(socket); !os.IsNotExist(err) {
		t.Fatalf("socket created for invalid token: %v", err)
	}
}
