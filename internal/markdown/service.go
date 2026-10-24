package markdown

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"reflect"

	rendererv1 "github.com/gotthboard/gotth-extension-markdown/proto/gotth/markdown/renderer/v1"
	extensioncore "github.com/gotthboard/gotth-extensions/pkg/extensions"
	extensionsv1 "github.com/gotthboard/gotth-sdk/gen/go/gotth/extensions/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type renderFunction func(context.Context, string) (Result, error)

type Service struct {
	extensionsv1.UnimplementedExtensionControlServer
	rendererv1.UnimplementedMarkdownRendererServer
	config              Config
	manifest            extensioncore.Manifest
	grant               extensioncore.Grant
	host                extensioncore.HostProfile
	session             extensioncore.Session
	tokenDigest         [sha256.Size]byte
	version             string
	configurationSHA256 string
	rendererIdentity    string
	semaphore           chan struct{}
	render              renderFunction
}

func NewService(runtime RuntimeConfig) (*Service, error) {
	negotiated, err := extensioncore.Negotiate(runtime.Manifest, runtime.Grant, runtime.Host)
	if err != nil || !reflect.DeepEqual(negotiated, runtime.Session) || !validRendererSession(runtime.Session) || !validToken(runtime.ServiceToken, 32, 256) || !validVersion(runtime.Version) || runtime.ConfigurationSHA256 == "" || runtime.Config.Profile != ProfileBBGFM || runtime.Config.MaxConcurrent < 1 || runtime.Config.MaxConcurrent > 8 {
		return nil, errors.New("complete Markdown runtime required")
	}
	manifestBytes, err := extensioncore.CanonicalManifest(runtime.Manifest)
	if err != nil {
		return nil, errors.New("complete Markdown runtime required")
	}
	manifest, err := extensioncore.ParseManifest(manifestBytes)
	if err != nil {
		return nil, errors.New("complete Markdown runtime required")
	}
	grantBytes, err := extensioncore.CanonicalGrant(runtime.Grant)
	if err != nil {
		return nil, errors.New("complete Markdown runtime required")
	}
	grant, err := extensioncore.ParseGrant(grantBytes)
	if err != nil {
		return nil, errors.New("complete Markdown runtime required")
	}
	host := extensioncore.HostProfile{Protocols: append([]extensioncore.VersionRange(nil), runtime.Host.Protocols...), Interfaces: append([]extensioncore.VersionRange(nil), runtime.Host.Interfaces...)}
	session, err := extensioncore.Negotiate(manifest, grant, host)
	if err != nil || !reflect.DeepEqual(session, runtime.Session) {
		return nil, errors.New("complete Markdown runtime required")
	}
	return &Service{
		config: runtime.Config, manifest: manifest, grant: grant, host: host,
		session: session, tokenDigest: sha256.Sum256([]byte(runtime.ServiceToken)), version: runtime.Version,
		configurationSHA256: runtime.ConfigurationSHA256,
		rendererIdentity:    RendererIdentity(runtime.Version, runtime.ConfigurationSHA256),
		semaphore:           make(chan struct{}, runtime.Config.MaxConcurrent), render: Render,
	}, nil
}

func (s *Service) authenticate(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "service identity required")
	}
	values := md.Get("x-gotth-service-token")
	var supplied string
	if len(values) == 1 {
		supplied = values[0]
	}
	suppliedDigest := sha256.Sum256([]byte(supplied))
	if len(values) != 1 || subtle.ConstantTimeCompare(suppliedDigest[:], s.tokenDigest[:]) != 1 {
		return status.Error(codes.Unauthenticated, "invalid service identity")
	}
	return nil
}

// UnaryAuthInterceptor rejects decoded unauthenticated RPCs before any
// application handler validates or otherwise processes the request.
func (s *Service) UnaryAuthInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if err := s.authenticate(ctx); err != nil {
		return nil, err
	}
	return handler(ctx, req)
}

func (s *Service) authorize(ctx context.Context, correlationID, capability, interfaceName string) error {
	if err := s.authenticate(ctx); err != nil {
		return err
	}
	if _, ok := ctx.Deadline(); !ok {
		return status.Error(codes.InvalidArgument, "deadline required")
	}
	if !validUUID(correlationID) {
		return status.Error(codes.InvalidArgument, "valid correlation ID required")
	}
	negotiated, err := extensioncore.Negotiate(s.manifest, s.grant, s.host)
	if err != nil || !reflect.DeepEqual(negotiated, s.session) || !validRendererSession(s.session) {
		return status.Error(codes.FailedPrecondition, "extension session invalid")
	}
	if capability != "" && !sessionAllowsCapability(s.session, capability) {
		return status.Error(codes.PermissionDenied, "capability not granted")
	}
	if interfaceName != "" && !sessionAllowsInterface(s.session, interfaceName, 1, 0) {
		return status.Error(codes.PermissionDenied, "interface not granted")
	}
	return nil
}

func sessionAllowsCapability(session extensioncore.Session, capability string) bool {
	for _, item := range session.Capabilities {
		if item == capability {
			return true
		}
	}
	return false
}

func sessionAllowsInterface(session extensioncore.Session, name string, major, minor uint32) bool {
	for _, item := range session.Interfaces {
		if item == (extensioncore.InterfaceGrant{Name: name, Major: major, Minor: minor}) {
			return true
		}
	}
	return false
}

func (s *Service) Handshake(ctx context.Context, in *extensionsv1.HandshakeRequest) (*extensionsv1.HandshakeResponse, error) {
	if err := s.authorize(ctx, in.GetCorrelationId(), "", ""); err != nil {
		return nil, err
	}
	if in == nil || len(in.GetChallenge()) != 32 {
		return nil, status.Error(codes.InvalidArgument, "valid extension handshake required")
	}
	if in.GetExpectedExtensionId() != s.session.ExtensionID || in.GetExpectedManifestSha256() != s.session.ManifestDigest || in.GetExpectedGrantSha256() != s.session.GrantDigest || in.GetExpectedSessionSha256() != s.session.Fingerprint {
		return nil, status.Error(codes.FailedPrecondition, "extension binding mismatch")
	}
	interfaces := make([]*extensionsv1.InterfaceVersion, len(s.session.Interfaces))
	for i, item := range s.session.Interfaces {
		interfaces[i] = &extensionsv1.InterfaceVersion{Name: item.Name, Major: item.Major, Minor: item.Minor}
	}
	return &extensionsv1.HandshakeResponse{
		Challenge: append([]byte(nil), in.GetChallenge()...), ExtensionId: s.session.ExtensionID,
		ExtensionVersion: s.session.ExtensionVersion, ManifestSha256: s.session.ManifestDigest,
		GrantSha256: s.session.GrantDigest, SessionSha256: s.session.Fingerprint,
		Control:    &extensionsv1.ProtocolVersion{Name: s.session.Control.Name, Major: s.session.Control.Major, Minor: s.session.Control.Minor},
		Interfaces: interfaces, Capabilities: append([]string(nil), s.session.Capabilities...),
	}, nil
}

func (s *Service) Health(ctx context.Context, in *extensionsv1.HealthRequest) (*extensionsv1.HealthResponse, error) {
	if err := s.authorize(ctx, in.GetCorrelationId(), "", ""); err != nil {
		return nil, err
	}
	if in == nil {
		return nil, status.Error(codes.InvalidArgument, "valid health request required")
	}
	if in.GetSessionSha256() != s.session.Fingerprint {
		return nil, status.Error(codes.FailedPrecondition, "extension session mismatch")
	}
	return &extensionsv1.HealthResponse{State: extensionsv1.HealthState_HEALTH_STATE_READY, Code: "extension.ready"}, nil
}

func (s *Service) Render(ctx context.Context, in *rendererv1.RenderRequest) (*rendererv1.RenderResponse, error) {
	if err := s.authorize(ctx, in.GetCorrelationId(), CapabilityRender, InterfaceName); err != nil {
		return nil, err
	}
	if in == nil {
		return nil, status.Error(codes.InvalidArgument, "valid render request required")
	}
	if in.GetProfile() != s.config.Profile {
		return nil, status.Error(codes.InvalidArgument, "unsupported rendering profile")
	}
	if len(in.GetSource()) == 0 || len(in.GetSource()) > MaximumSourceBytes {
		return nil, status.Error(codes.InvalidArgument, "invalid Markdown source")
	}
	renderContext, cancel := context.WithTimeout(ctx, s.config.Timeout())
	defer cancel()
	select {
	case s.semaphore <- struct{}{}:
		defer func() { <-s.semaphore }()
	case <-renderContext.Done():
		return nil, renderContextStatus(ctx, renderContext)
	default:
		return nil, status.Error(codes.ResourceExhausted, "render capacity exhausted")
	}
	result, err := s.render(renderContext, in.GetSource())
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return nil, status.Error(codes.Canceled, "render canceled")
		case errors.Is(err, context.DeadlineExceeded):
			return nil, status.Error(codes.DeadlineExceeded, "render deadline exceeded")
		case errors.Is(err, ErrInvalidSource), errors.Is(err, ErrNoRenderableContent):
			return nil, status.Error(codes.InvalidArgument, "invalid Markdown source")
		case errors.Is(err, ErrOutputTooLarge):
			return nil, status.Error(codes.ResourceExhausted, "rendered output exceeds limit")
		default:
			return nil, status.Error(codes.Internal, "render failed")
		}
	}
	if result.HTML == "" {
		return nil, status.Error(codes.Internal, "incomplete render rejected")
	}
	return &rendererv1.RenderResponse{Schema: ResponseSchema, Complete: true, Html: result.HTML, VisibleText: result.VisibleText, RendererIdentitySha256: s.rendererIdentity, Engine: EngineIdentity, Profile: ProfileBBGFM, Sanitizer: SanitizerIdentity, ConfigurationSha256: s.configurationSHA256, ExtensionVersion: s.version}, nil
}

func renderContextStatus(parent, renderContext context.Context) error {
	if errors.Is(parent.Err(), context.Canceled) || errors.Is(renderContext.Err(), context.Canceled) {
		return status.Error(codes.Canceled, "render canceled")
	}
	return status.Error(codes.DeadlineExceeded, "render deadline exceeded")
}
