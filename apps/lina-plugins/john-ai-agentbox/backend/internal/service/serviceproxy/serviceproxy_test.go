// This file verifies service-proxy access boundaries. Runtime-backed service
// data is intentionally unavailable in this migration slice, but invisible
// Agents and services must still be rejected before runtime state is considered.

package serviceproxy

import (
	"context"
	"testing"
	"time"

	"lina-core/pkg/bizerr"

	accesssvc "john-ai-agentbox/backend/internal/service/access"
)

// TestNewRequiresAccessService verifies constructor dependency validation.
func TestNewRequiresAccessService(t *testing.T) {
	if _, err := New(nil, nil); err == nil {
		t.Fatal("expected constructor to reject nil access service")
	}
}

// TestServiceRejectsInvisibleAgentBeforeRuntime verifies service lookup checks ownership first.
func TestServiceRejectsInvisibleAgentBeforeRuntime(t *testing.T) {
	service, err := New(&fakeAccessService{err: bizerr.NewCode(accesssvc.CodeAccessResourceUnavailable)}, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.Service(context.Background(), "usr-owner", "agt-other", "svc-secret")
	if !bizerr.Is(err, accesssvc.CodeAccessResourceUnavailable) {
		t.Fatalf("expected invisible resource error, got %v", err)
	}
}

// TestServicesRejectInvisibleAgentBeforeRuntime verifies list calls check Agent ownership.
func TestServicesRejectInvisibleAgentBeforeRuntime(t *testing.T) {
	service, err := New(&fakeAccessService{err: bizerr.NewCode(accesssvc.CodeAccessResourceUnavailable)}, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.Services(context.Background(), "usr-owner", "agt-other")
	if !bizerr.Is(err, accesssvc.CodeAccessResourceUnavailable) {
		t.Fatalf("expected invisible resource error, got %v", err)
	}
}

// TestCreateBridgeRejectsInvisibleServiceBeforeRuntime verifies bridge creation checks service visibility.
func TestCreateBridgeRejectsInvisibleServiceBeforeRuntime(t *testing.T) {
	service, err := New(&fakeAccessService{err: bizerr.NewCode(accesssvc.CodeAccessResourceUnavailable)}, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.CreateServiceBridge(context.Background(), "usr-owner", "agt-other", BridgeInput{
		ServiceID:     "svc-secret",
		ListenAddress: "127.0.0.1",
		Port:          3000,
	})
	if !bizerr.Is(err, accesssvc.CodeAccessResourceUnavailable) {
		t.Fatalf("expected invisible resource error, got %v", err)
	}
}

// TestVisibleServiceWithoutRuntimeReturnsRuntimeUnavailable verifies nil runtime degradation.
func TestVisibleServiceWithoutRuntimeReturnsRuntimeUnavailable(t *testing.T) {
	service, err := New(&fakeAccessService{}, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.Service(context.Background(), "usr-owner", "agt-owned", "svc-owned")
	if !bizerr.Is(err, CodeServiceProxyRuntimeUnavailable) {
		t.Fatalf("expected runtime unavailable error, got %v", err)
	}
}

// TestServicesUsesRuntimeBackend verifies service discovery delegates after ownership checks.
func TestServicesUsesRuntimeBackend(t *testing.T) {
	backend := &fakeRuntimeBackend{
		services: []RuntimeServiceInfo{
			{
				ID:            "svc-3000",
				AgentID:       "agt-owned",
				Port:          3000,
				Protocol:      AgentServiceProtocolUnknown,
				AccessStatus:  AgentServiceAccessDirect,
				LastCheckedAt: time.Unix(10, 0).UnixMilli(),
			},
		},
	}
	service, err := New(&fakeAccessService{}, backend)
	if err != nil {
		t.Fatal(err)
	}

	items, err := service.Services(context.Background(), " usr-owner ", " agt-owned ")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != "svc-3000" || items[0].ProxyURL != "" || items[0].TunnelURL != "" {
		t.Fatalf("unexpected services response: %#v", items)
	}
	if backend.lastUserID != "usr-owner" || backend.lastAgentID != "agt-owned" {
		t.Fatalf("unexpected runtime scope: user=%q agent=%q", backend.lastUserID, backend.lastAgentID)
	}
}

// TestServiceFiltersRuntimeBackendResult verifies service lookup does not expose unrelated service IDs.
func TestServiceFiltersRuntimeBackendResult(t *testing.T) {
	service, err := New(&fakeAccessService{}, &fakeRuntimeBackend{
		services: []RuntimeServiceInfo{{ID: "svc-3000", AgentID: "agt-owned", Port: 3000}},
	})
	if err != nil {
		t.Fatal(err)
	}

	item, err := service.Service(context.Background(), "usr-owner", "agt-owned", "svc-3000")
	if err != nil {
		t.Fatal(err)
	}
	if item.ID != "svc-3000" {
		t.Fatalf("unexpected service response: %#v", item)
	}
	_, err = service.Service(context.Background(), "usr-owner", "agt-owned", "svc-missing")
	if !bizerr.Is(err, CodeServiceProxyRuntimeUnavailable) {
		t.Fatalf("expected unavailable for missing runtime service, got %v", err)
	}
}

// TestRejectsInvalidBridgeInput verifies invalid bridge inputs fail before access checks.
func TestRejectsInvalidBridgeInput(t *testing.T) {
	access := &fakeAccessService{}
	service, err := New(access, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.CreateServiceBridge(context.Background(), "usr-owner", "agt-owned", BridgeInput{
		ServiceID:     "svc-owned",
		ListenAddress: "0.0.0.0",
		Port:          3000,
	})
	if !bizerr.Is(err, CodeServiceProxyInvalidInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
	if access.agentChecks != 0 || access.proxyChecks != 0 {
		t.Fatalf("expected invalid bridge input to avoid access checks, got agent=%d proxy=%d", access.agentChecks, access.proxyChecks)
	}
}

type fakeAccessService struct {
	accesssvc.Service
	err         error
	agentChecks int
	proxyChecks int
}

func (s *fakeAccessService) EnsureAgentVisible(context.Context, string, string) error {
	s.agentChecks++
	return s.err
}

func (s *fakeAccessService) EnsureServiceProxyVisible(context.Context, string, string, string) error {
	s.proxyChecks++
	return s.err
}

type fakeRuntimeBackend struct {
	lastUserID  string
	lastAgentID string
	services    []RuntimeServiceInfo
	err         error
}

func (b *fakeRuntimeBackend) RuntimeServices(_ context.Context, userID string, agentID string) ([]RuntimeServiceInfo, error) {
	b.lastUserID, b.lastAgentID = userID, agentID
	if b.err != nil {
		return nil, b.err
	}
	return b.services, nil
}
