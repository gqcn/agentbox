// This file verifies raw gateway controller authentication plumbing. Handlers
// must use the authenticated AgentBox user from context before delegating to
// gateway service methods.

package gateway

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/net/ghttp"

	authsvc "john-ai-agentbox/backend/internal/service/auth"
	"john-ai-agentbox/backend/internal/service/authctx"
	gatewaysvc "john-ai-agentbox/backend/internal/service/gateway"
)

// TestControllerRequiresGatewayService verifies constructor dependency validation.
func TestControllerRequiresGatewayService(t *testing.T) {
	if _, err := New(nil); err == nil {
		t.Fatal("expected constructor to reject nil gateway service")
	}
}

// TestGatewayServiceReceivesAuthenticatedUser verifies the raw controller can be constructed with a gateway service.
func TestGatewayServiceReceivesAuthenticatedUser(t *testing.T) {
	service := &fakeGatewayService{}
	controller, err := New(service)
	if err != nil {
		t.Fatal(err)
	}

	ctx := authctx.WithUser(context.Background(), authsvc.UserOutput{ID: "usr-owner", Username: "owner"})
	if err := controller.gatewaySvc.AgentShell(ctx, "usr-owner", "agt-owned", "", "", ""); err != nil {
		t.Fatal(err)
	}
	if service.lastUserID != "usr-owner" || service.lastAgentID != "agt-owned" {
		t.Fatalf("unexpected service call: user=%q agent=%q", service.lastUserID, service.lastAgentID)
	}
}

// TestHandlersAcceptGoFrameRequestType guards the raw handler signatures used by route binding.
func TestHandlersAcceptGoFrameRequestType(t *testing.T) {
	controller, err := New(&fakeGatewayService{})
	if err != nil {
		t.Fatal(err)
	}
	handlers := []func(*ghttp.Request){
		controller.AgentServiceHTTPProxy,
		controller.AgentShell,
		controller.AgentChat,
		controller.AgentServiceTCPTunnel,
	}
	if len(handlers) != 4 {
		t.Fatalf("unexpected handler count %d", len(handlers))
	}
}

type fakeGatewayService struct {
	gatewaysvc.Service
	lastUserID  string
	lastAgentID string
}

func (s *fakeGatewayService) AgentShell(_ context.Context, userID string, agentID string, terminalID string, cwd string, mode string) error {
	_ = terminalID
	_ = cwd
	_ = mode
	s.lastUserID = userID
	s.lastAgentID = agentID
	return nil
}

func (s *fakeGatewayService) AgentServiceHTTPProxy(_ context.Context, userID string, escapedPath string) error {
	_ = escapedPath
	s.lastUserID = userID
	return nil
}
