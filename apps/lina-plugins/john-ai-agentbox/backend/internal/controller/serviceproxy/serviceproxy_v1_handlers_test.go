// This file verifies service-proxy controller ownership plumbing. The
// controller must use the authenticated AgentBox user from context for all
// service-proxy calls and reject calls without plugin authentication context.

package serviceproxy

import (
	"context"
	"testing"

	"lina-core/pkg/bizerr"

	v1 "john-ai-agentbox/backend/api/serviceproxy/v1"
	authsvc "john-ai-agentbox/backend/internal/service/auth"
	"john-ai-agentbox/backend/internal/service/authctx"
	serviceproxysvc "john-ai-agentbox/backend/internal/service/serviceproxy"
)

// TestServiceUsesAuthenticatedUser verifies service lookup is user scoped.
func TestServiceUsesAuthenticatedUser(t *testing.T) {
	service := &fakeServiceProxyService{
		service: serviceproxysvc.RuntimeServiceInfo{
			ID:            "svc-owned",
			AgentID:       "agt-test",
			Port:          3000,
			Protocol:      serviceproxysvc.AgentServiceProtocolHTTP,
			AccessStatus:  serviceproxysvc.AgentServiceAccessDirect,
			LastCheckedAt: 1704067200000,
		},
	}
	controller := newTestController(t, service)
	ctx := authctx.WithUser(context.Background(), authsvc.UserOutput{ID: "usr-owner", Username: "owner"})

	res, err := controller.Service(ctx, &v1.ServiceReq{ID: "agt-test", ServiceID: "svc-owned"})
	if err != nil {
		t.Fatal(err)
	}
	if service.lastUserID != "usr-owner" || service.lastAgentID != "agt-test" || service.lastServiceID != "svc-owned" {
		t.Fatalf("unexpected service call: user=%q agent=%q service=%q", service.lastUserID, service.lastAgentID, service.lastServiceID)
	}
	if res.ID != "svc-owned" || res.Port != 3000 {
		t.Fatalf("unexpected service response: %#v", res)
	}
}

// TestCreateBridgeUsesAuthenticatedUser verifies bridge creation is user scoped.
func TestCreateBridgeUsesAuthenticatedUser(t *testing.T) {
	service := &fakeServiceProxyService{
		bridge: serviceproxysvc.BridgeInfo{ID: "brg-owned", AgentID: "agt-test", ServiceID: "svc-owned", Port: 3000},
	}
	controller := newTestController(t, service)
	ctx := authctx.WithUser(context.Background(), authsvc.UserOutput{ID: "usr-owner", Username: "owner"})

	res, err := controller.CreateServiceBridge(ctx, &v1.CreateServiceBridgeReq{
		ID:            "agt-test",
		ServiceID:     "svc-owned",
		ListenAddress: "127.0.0.1",
		Port:          3000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if service.lastUserID != "usr-owner" || service.lastAgentID != "agt-test" || service.lastInput.ServiceID != "svc-owned" {
		t.Fatalf("unexpected bridge call: user=%q agent=%q input=%#v", service.lastUserID, service.lastAgentID, service.lastInput)
	}
	if res.ID != "brg-owned" {
		t.Fatalf("unexpected bridge response: %#v", res)
	}
}

// TestServicesRequiresAuthenticatedUser verifies missing auth is rejected.
func TestServicesRequiresAuthenticatedUser(t *testing.T) {
	controller := newTestController(t, &fakeServiceProxyService{})

	_, err := controller.Services(context.Background(), &v1.ServicesReq{ID: "agt-test"})
	if !bizerr.Is(err, authsvc.CodeAuthRequired) {
		t.Fatalf("expected auth required error, got %v", err)
	}
}

func newTestController(t *testing.T, serviceProxySvc serviceproxysvc.Service) *ControllerV1 {
	t.Helper()
	controller, err := NewV1(serviceProxySvc)
	if err != nil {
		t.Fatal(err)
	}
	typed, ok := controller.(*ControllerV1)
	if !ok {
		t.Fatalf("unexpected controller type %T", controller)
	}
	return typed
}

type fakeServiceProxyService struct {
	lastUserID    string
	lastAgentID   string
	lastServiceID string
	lastBridgeID  string
	lastInput     serviceproxysvc.BridgeInput
	services      []serviceproxysvc.RuntimeServiceInfo
	service       serviceproxysvc.RuntimeServiceInfo
	bridges       []serviceproxysvc.BridgeInfo
	bridge        serviceproxysvc.BridgeInfo
	err           error
}

func (s *fakeServiceProxyService) Services(_ context.Context, userID string, agentID string) ([]serviceproxysvc.RuntimeServiceInfo, error) {
	s.lastUserID, s.lastAgentID = userID, agentID
	if s.err != nil {
		return nil, s.err
	}
	return s.services, nil
}

func (s *fakeServiceProxyService) Service(_ context.Context, userID string, agentID string, serviceID string) (serviceproxysvc.RuntimeServiceInfo, error) {
	s.lastUserID, s.lastAgentID, s.lastServiceID = userID, agentID, serviceID
	if s.err != nil {
		return serviceproxysvc.RuntimeServiceInfo{}, s.err
	}
	return s.service, nil
}

func (s *fakeServiceProxyService) ServiceBridges(_ context.Context, userID string, agentID string) ([]serviceproxysvc.BridgeInfo, error) {
	s.lastUserID, s.lastAgentID = userID, agentID
	if s.err != nil {
		return nil, s.err
	}
	return s.bridges, nil
}

func (s *fakeServiceProxyService) CreateServiceBridge(_ context.Context, userID string, agentID string, input serviceproxysvc.BridgeInput) (serviceproxysvc.BridgeInfo, error) {
	s.lastUserID, s.lastAgentID, s.lastInput = userID, agentID, input
	if s.err != nil {
		return serviceproxysvc.BridgeInfo{}, s.err
	}
	return s.bridge, nil
}

func (s *fakeServiceProxyService) DeleteServiceBridge(_ context.Context, userID string, agentID string, bridgeID string) (bool, error) {
	s.lastUserID, s.lastAgentID, s.lastBridgeID = userID, agentID, bridgeID
	if s.err != nil {
		return false, s.err
	}
	return true, nil
}
