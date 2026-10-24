package markdown

import (
	"context"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	rendererv1 "github.com/gotthboard/gotth-extension-markdown/proto/gotth/markdown/renderer/v1"
	extensioncore "github.com/gotthboard/gotth-extensions/pkg/extensions"
	extensionsv1 "github.com/gotthboard/gotth-sdk/gen/go/gotth/extensions/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const testToken = "12345678901234567890123456789012"
const testDigest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func testRuntime() RuntimeConfig {
	manifest, err := Manifest("dev")
	if err != nil {
		panic(err)
	}
	manifestDigest, err := extensioncore.ManifestDigest(manifest)
	if err != nil {
		panic(err)
	}
	grant := extensioncore.Grant{Schema: extensioncore.GrantSchema, InstanceID: "11111111-1111-1111-1111-111111111111", ExtensionID: ExtensionID, ManifestDigest: manifestDigest, Capabilities: []string{CapabilityRender}, Interfaces: []extensioncore.InterfaceGrant{{Name: InterfaceName, Major: 1, Minor: 0}}, Secrets: []string{}}
	host := MarkdownHostProfile()
	session, err := extensioncore.Negotiate(manifest, grant, host)
	if err != nil {
		panic(err)
	}
	return RuntimeConfig{Config: Config{Schema: ConfigSchema, Profile: ProfileBBGFM, TimeoutMilliseconds: 100, MaxConcurrent: 2}, Manifest: manifest, Grant: grant, Host: host, Session: session, ServiceToken: testToken, Version: "dev", ConfigurationSHA256: testDigest}
}
func authorizedContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	t.Cleanup(cancel)
	return metadata.NewIncomingContext(ctx, metadata.Pairs("x-gotth-service-token", testToken))
}

func TestServiceAuthenticationControlAndRender(t *testing.T) {
	s, err := NewService(testRuntime())
	if err != nil {
		t.Fatal(err)
	}
	request := &rendererv1.RenderRequest{CorrelationId: "22222222-2222-2222-2222-222222222222", Profile: ProfileBBGFM, Source: "Hello **world**"}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := s.Render(ctx, request); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("unauthenticated code = %v: %v", status.Code(err), err)
	}
	response, err := s.Render(authorizedContext(t), request)
	if err != nil {
		t.Fatal(err)
	}
	if !response.GetComplete() || response.GetHtml() != "<p>Hello <strong>world</strong></p>\n" || response.GetVisibleText() != "Hello world" || response.GetRendererIdentitySha256() == "" || response.GetConfigurationSha256() != testDigest || response.GetExtensionVersion() != "dev" {
		t.Fatalf("response = %+v", response)
	}
	health, err := s.Health(authorizedContext(t), &extensionsv1.HealthRequest{CorrelationId: "33333333-3333-3333-3333-333333333333", SessionSha256: s.session.Fingerprint})
	if err != nil || health.GetState() != extensionsv1.HealthState_HEALTH_STATE_READY {
		t.Fatalf("health=%+v err=%v", health, err)
	}
}

func TestServiceAllowsSemanticDocumentWithEmptyVisibleText(t *testing.T) {
	s, err := NewService(testRuntime())
	if err != nil {
		t.Fatal(err)
	}
	response, err := s.Render(authorizedContext(t), &rendererv1.RenderRequest{CorrelationId: "77777777-7777-4777-8777-777777777777", Profile: ProfileBBGFM, Source: "---"})
	if err != nil {
		t.Fatal(err)
	}
	if !response.GetComplete() || response.GetHtml() != "<hr>\n" || response.GetVisibleText() != "" {
		t.Fatalf("semantic document response = %+v", response)
	}
}

func TestServiceRejectsMalformedBoundary(t *testing.T) {
	s, _ := NewService(testRuntime())
	valid := &rendererv1.RenderRequest{CorrelationId: "22222222-2222-2222-2222-222222222222", Profile: ProfileBBGFM, Source: "body"}
	tests := []struct {
		name    string
		request *rendererv1.RenderRequest
		code    codes.Code
	}{
		{"nil", nil, codes.InvalidArgument},
		{"correlation", &rendererv1.RenderRequest{CorrelationId: "bad", Profile: ProfileBBGFM, Source: "body"}, codes.InvalidArgument},
		{"profile", &rendererv1.RenderRequest{CorrelationId: valid.CorrelationId, Profile: "gotth.markdown.rich.v1", Source: "body"}, codes.InvalidArgument},
		{"source", &rendererv1.RenderRequest{CorrelationId: valid.CorrelationId, Profile: ProfileBBGFM, Source: strings.Repeat("x", MaximumSourceBytes+1)}, codes.InvalidArgument},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := s.Render(authorizedContext(t), tt.request)
			if status.Code(err) != tt.code {
				t.Fatalf("code=%v err=%v", status.Code(err), err)
			}
		})
	}
	withoutDeadline := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-gotth-service-token", testToken))
	if _, err := s.Render(withoutDeadline, valid); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("missing deadline code=%v", status.Code(err))
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("x-gotth-service-token", testToken, "x-gotth-service-token", testToken))
	if _, err := s.Render(ctx, valid); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("duplicate token code=%v", status.Code(err))
	}
}

func TestServiceErrorsDoNotEchoPrivateInput(t *testing.T) {
	s, _ := NewService(testRuntime())
	const private = "private-source-marker-9917"
	_, err := s.Render(authorizedContext(t), &rendererv1.RenderRequest{CorrelationId: "22222222-2222-2222-2222-222222222222", Profile: ProfileBBGFM, Source: "<script>" + private + "</script>"})
	if err == nil || strings.Contains(err.Error(), private) || strings.Contains(err.Error(), testToken) {
		t.Fatalf("unredacted error: %v", err)
	}
}

func TestServiceTimeoutCancelsRender(t *testing.T) {
	runtimeConfig := testRuntime()
	runtimeConfig.Config.TimeoutMilliseconds = 50
	s, _ := NewService(runtimeConfig)
	s.render = func(ctx context.Context, _ string) (Result, error) { <-ctx.Done(); return Result{}, ctx.Err() }
	_, err := s.Render(authorizedContext(t), &rendererv1.RenderRequest{CorrelationId: "22222222-2222-2222-2222-222222222222", Profile: ProfileBBGFM, Source: "body"})
	if status.Code(err) != codes.DeadlineExceeded {
		t.Fatalf("timeout code=%v err=%v", status.Code(err), err)
	}
}

func TestServiceRejectsWaitingRendersAcrossConnections(t *testing.T) {
	runtimeConfig := testRuntime()
	runtimeConfig.Config.MaxConcurrent = 2
	runtimeConfig.Config.TimeoutMilliseconds = 1000
	s, _ := NewService(runtimeConfig)
	entered := make(chan struct{}, 2)
	release := make(chan struct{})
	s.render = func(context.Context, string) (Result, error) {
		entered <- struct{}{}
		<-release
		return Result{HTML: "<p>x</p>\n", VisibleText: "x"}, nil
	}
	listener := bufconn.Listen(1 << 20)
	server := grpc.NewServer(grpc.UnaryInterceptor(s.UnaryAuthInterceptor))
	rendererv1.RegisterMarkdownRendererServer(server, s)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() { server.Stop(); _ = listener.Close() })
	clients := make([]rendererv1.MarkdownRendererClient, 3)
	for i := range clients {
		conn, err := grpc.NewClient("passthrough:///renderer", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = conn.Close() })
		clients[i] = rendererv1.NewMarkdownRendererClient(conn)
	}
	request := &rendererv1.RenderRequest{CorrelationId: "22222222-2222-2222-2222-222222222222", Profile: ProfileBBGFM, Source: "x"}
	call := func(client rendererv1.MarkdownRendererClient) error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		ctx = metadata.AppendToOutgoingContext(ctx, "x-gotth-service-token", testToken)
		_, err := client.Render(ctx, request)
		return err
	}
	var wg sync.WaitGroup
	errors := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(client rendererv1.MarkdownRendererClient) {
			defer wg.Done()
			errors <- call(client)
		}(clients[i])
	}
	<-entered
	<-entered
	if err := call(clients[2]); status.Code(err) != codes.ResourceExhausted {
		t.Fatalf("third connection admission code=%v err=%v", status.Code(err), err)
	}
	close(release)
	wg.Wait()
	close(errors)
	for err := range errors {
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestUnaryAuthInterceptorRejectsBeforeApplicationHandler(t *testing.T) {
	s, _ := NewService(testRuntime())
	called := false
	handler := func(context.Context, any) (any, error) {
		called = true
		return nil, status.Error(codes.InvalidArgument, "application validation")
	}
	if _, err := s.UnaryAuthInterceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, handler); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("unauthenticated interceptor code=%v err=%v", status.Code(err), err)
	}
	if called {
		t.Fatal("application handler ran before authentication")
	}
}

func TestControlRejectsMalformedAndMismatchedBindings(t *testing.T) {
	s, _ := NewService(testRuntime())
	if _, err := s.Health(authorizedContext(t), nil); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("nil health code=%v", status.Code(err))
	}
	if _, err := s.Health(authorizedContext(t), &extensionsv1.HealthRequest{CorrelationId: "33333333-3333-3333-3333-333333333333", SessionSha256: strings.Repeat("b", 64)}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("health mismatch code=%v", status.Code(err))
	}
	if _, err := s.Handshake(authorizedContext(t), &extensionsv1.HandshakeRequest{CorrelationId: "22222222-2222-2222-2222-222222222222", Challenge: []byte("short")}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("short challenge code=%v", status.Code(err))
	}
	request := &extensionsv1.HandshakeRequest{CorrelationId: "22222222-2222-2222-2222-222222222222", Challenge: []byte("01234567890123456789012345678901"), ExpectedExtensionId: s.session.ExtensionID, ExpectedManifestSha256: s.session.ManifestDigest, ExpectedGrantSha256: s.session.GrantDigest, ExpectedSessionSha256: s.session.Fingerprint}
	request.ExpectedExtensionId = "gotth.render.other"
	if _, err := s.Handshake(authorizedContext(t), request); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("handshake mismatch code=%v", status.Code(err))
	}
}

func TestHandshakeEchoesExactPublishedBoundary(t *testing.T) {
	runtimeConfig := testRuntime()
	s, _ := NewService(runtimeConfig)
	challenge := []byte("01234567890123456789012345678901")
	response, err := s.Handshake(authorizedContext(t), &extensionsv1.HandshakeRequest{CorrelationId: "22222222-2222-2222-2222-222222222222", Challenge: challenge, ExpectedExtensionId: s.session.ExtensionID, ExpectedManifestSha256: s.session.ManifestDigest, ExpectedGrantSha256: s.session.GrantDigest, ExpectedSessionSha256: s.session.Fingerprint})
	if err != nil {
		t.Fatal(err)
	}
	if string(response.GetChallenge()) != string(challenge) || response.GetControl().GetName() != "gotth.extensions.control" || len(response.GetInterfaces()) != 1 || response.GetInterfaces()[0].GetName() != InterfaceName {
		t.Fatalf("handshake=%+v", response)
	}
}

func TestGRPCServerObservesCallerCancellation(t *testing.T) {
	runtimeConfig := testRuntime()
	runtimeConfig.Config.MaxConcurrent = 1
	runtimeConfig.Config.TimeoutMilliseconds = 1000
	s, err := NewService(runtimeConfig)
	if err != nil {
		t.Fatal(err)
	}
	entered := make(chan struct{})
	serverObserved := make(chan struct{})
	s.render = func(ctx context.Context, _ string) (Result, error) {
		close(entered)
		<-ctx.Done()
		close(serverObserved)
		return Result{}, ctx.Err()
	}
	listener := bufconn.Listen(1 << 20)
	server := grpc.NewServer(grpc.UnaryInterceptor(s.UnaryAuthInterceptor))
	rendererv1.RegisterMarkdownRendererServer(server, s)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() { server.Stop(); _ = listener.Close() })
	conn, err := grpc.NewClient("passthrough:///renderer", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	client := rendererv1.NewMarkdownRendererClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	t.Cleanup(cancel)
	ctx = metadata.AppendToOutgoingContext(ctx, "x-gotth-service-token", testToken)
	done := make(chan error, 1)
	go func() {
		_, callErr := client.Render(ctx, &rendererv1.RenderRequest{CorrelationId: "33333333-3333-3333-3333-333333333333", Profile: ProfileBBGFM, Source: "body"})
		done <- callErr
	}()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("render handler was not entered")
	}
	cancel()
	select {
	case <-serverObserved:
	case <-time.After(time.Second):
		t.Fatal("server render context did not observe caller cancellation")
	}
	select {
	case callErr := <-done:
		if status.Code(callErr) != codes.Canceled {
			t.Fatalf("cancellation code=%v err=%v", status.Code(callErr), callErr)
		}
	case <-time.After(time.Second):
		t.Fatal("canceled RPC did not return")
	}
}

func TestServiceRejectsMutatedNegotiatedAuthority(t *testing.T) {
	s, err := NewService(testRuntime())
	if err != nil {
		t.Fatal(err)
	}
	s.session.Capabilities[0] = "markdown.other"
	_, err = s.Render(authorizedContext(t), &rendererv1.RenderRequest{CorrelationId: "22222222-2222-2222-2222-222222222222", Profile: ProfileBBGFM, Source: "body"})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("mutated authority code=%v err=%v", status.Code(err), err)
	}
}
