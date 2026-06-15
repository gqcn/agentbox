// This file verifies Agent runtime helper behavior that does not require
// database state. Database-backed lifecycle wiring is covered by package and
// route compile gates.

package catalog

import "testing"

// TestNormalizeAgentRuntimeStatus verifies only stable runtime states are persisted.
func TestNormalizeAgentRuntimeStatus(t *testing.T) {
	if normalizeAgentRuntimeStatus(AgentRuntimeStatusRunning) != AgentRuntimeStatusRunning {
		t.Fatal("expected running status to stay running")
	}
	if normalizeAgentRuntimeStatus("unexpected") != AgentRuntimeStatusStopped {
		t.Fatal("expected unexpected runtime status to normalize to stopped")
	}
}

// TestCatalogNewKeepsNilRuntimeBackend verifies runtime lifecycle remains unavailable unless startup injects a backend.
func TestCatalogNewKeepsNilRuntimeBackend(t *testing.T) {
	service, err := New(nil, Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	impl, ok := service.(*serviceImpl)
	if !ok {
		t.Fatalf("service type = %T", service)
	}
	if impl.runtimeBackend != nil {
		t.Fatal("expected nil runtime backend when config omits it")
	}
}
