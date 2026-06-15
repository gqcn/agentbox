// This file verifies AgentBox raw gateway access boundaries. Runtime-backed
// sockets are intentionally unavailable in this migration slice, but invisible
// Agents, Chat sessions, and services must be rejected first.

package gateway

import (
	"context"
	"testing"

	"lina-core/pkg/bizerr"

	accesssvc "john-ai-agentbox/backend/internal/service/access"
)

// TestAgentChatRejectsInvisibleSessionBeforeRuntime verifies Chat ownership is checked first.
func TestAgentChatRejectsInvisibleSessionBeforeRuntime(t *testing.T) {
	service, err := New(&fakeAccessService{err: bizerr.NewCode(accesssvc.CodeAccessResourceUnavailable)})
	if err != nil {
		t.Fatal(err)
	}

	err = service.AgentChat(context.Background(), "usr-owner", "agt-other", "chat-other", "/home/agent/workspace")
	if !bizerr.Is(err, accesssvc.CodeAccessResourceUnavailable) {
		t.Fatalf("expected invisible resource error, got %v", err)
	}
}

// TestVisibleGatewayEntrypointsReturnRuntimeUnavailable verifies visible raw routes do not fake sockets.
func TestVisibleGatewayEntrypointsReturnRuntimeUnavailable(t *testing.T) {
	service, err := New(&fakeAccessService{})
	if err != nil {
		t.Fatal(err)
	}

	checks := []struct {
		name string
		run  func() error
	}{
		{
			name: "http-proxy",
			run: func() error {
				return service.AgentServiceHTTPProxy(context.Background(), "usr-owner", "/x/john-ai-agentbox/api/v1/proxy/opaque-key/app/index.html")
			},
		},
		{
			name: "shell",
			run: func() error {
				return service.AgentShell(context.Background(), "usr-owner", "agt-owned", "", "/home/agent/workspace", "")
			},
		},
		{
			name: "chat",
			run: func() error {
				return service.AgentChat(context.Background(), "usr-owner", "agt-owned", "chat-owned", "/home/agent/workspace")
			},
		},
		{
			name: "tunnel",
			run: func() error {
				return service.AgentServiceTCPTunnel(context.Background(), "usr-owner", "agt-owned", "svc-owned", "opaque-key")
			},
		},
	}
	for _, check := range checks {
		if err := check.run(); !bizerr.Is(err, CodeGatewayRuntimeUnavailable) {
			t.Fatalf("%s: expected runtime unavailable error, got %v", check.name, err)
		}
	}
}

// TestParseServiceProxyPathScopesKeyAndPath verifies plugin proxy paths keep
// the opaque access key separate from the upstream relative path.
func TestParseServiceProxyPathScopesKeyAndPath(t *testing.T) {
	key, restPath, err := parseServiceProxyPath("/x/john-ai-agentbox/api/v1/proxy/key%2Fencoded/app%20space/index.html")
	if err != nil {
		t.Fatal(err)
	}
	if key != "key/encoded" || restPath != "app space/index.html" {
		t.Fatalf("key=%q restPath=%q, want decoded key and path", key, restPath)
	}
	if _, _, err := parseServiceProxyPath("/x/john-ai-agentbox/api/v1/proxy/"); !bizerr.Is(err, CodeGatewayInvalidInput) {
		t.Fatalf("expected invalid empty key, got %v", err)
	}
	if _, _, err := parseServiceProxyPath("/x/john-ai-agentbox/api/v1/services/key"); !bizerr.Is(err, CodeGatewayInvalidInput) {
		t.Fatalf("expected invalid prefix, got %v", err)
	}
}

type fakeAccessService struct {
	accesssvc.Service
	err error
}

func (s *fakeAccessService) EnsureAgentVisible(context.Context, string, string) error {
	return s.err
}

func (s *fakeAccessService) EnsureChatSessionVisible(context.Context, string, string, string) error {
	return s.err
}

func (s *fakeAccessService) EnsureServiceProxyVisible(context.Context, string, string, string) error {
	return s.err
}

func (s *fakeAccessService) EnsureTerminalSessionVisible(context.Context, string, string, string) error {
	return s.err
}
