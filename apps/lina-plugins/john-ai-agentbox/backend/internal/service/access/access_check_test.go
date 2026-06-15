// This file verifies AgentBox access-check rejection semantics without relying
// on shared database state. Invisible resources must be rejected without
// leaking names, status, counts, shell, paths, or proxy identifiers.

package access

import (
	"context"
	"errors"
	"strings"
	"testing"

	"lina-core/pkg/bizerr"
)

// TestServiceRejectsInvisibleResourcesWithoutMetadataLeaks verifies false
// ownership checks all collapse to one non-leaking resource-unavailable error.
func TestServiceRejectsInvisibleResourcesWithoutMetadataLeaks(t *testing.T) {
	store := &accessTestStore{}
	service, err := New(store)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	checks := []struct {
		name string
		run  func() error
	}{
		{name: "agent", run: func() error {
			return service.EnsureAgentVisible(ctx, "usr-current", "agt-secret")
		}},
		{name: "chat", run: func() error {
			return service.EnsureChatSessionVisible(ctx, "usr-current", "agt-secret", "chat-secret")
		}},
		{name: "terminal", run: func() error {
			return service.EnsureTerminalSessionVisible(ctx, "usr-current", "agt-secret", "term-secret")
		}},
		{name: "workspace", run: func() error {
			return service.EnsureWorkspaceResourceVisible(ctx, "usr-current", "agt-secret", "/home/agent/workspace/private.log")
		}},
		{name: "proxy", run: func() error {
			return service.EnsureServiceProxyVisible(ctx, "usr-current", "agt-secret", "svc-secret")
		}},
	}

	for _, tt := range checks {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.run()
			if !bizerr.Is(err, CodeAccessResourceUnavailable) {
				t.Fatalf("expected resource unavailable error, got %v", err)
			}
			assertAccessErrorHasNoLeak(t, err)
		})
	}
}

// TestServiceWrapsStoreFailures verifies backend failures use a structured
// store-unavailable code instead of leaking low-level database text.
func TestServiceWrapsStoreFailures(t *testing.T) {
	service, err := New(&accessTestStore{err: errors.New("database leaked other-user-agent-name")})
	if err != nil {
		t.Fatal(err)
	}

	err = service.EnsureAgentVisible(context.Background(), "usr-current", "agt-secret")
	if !bizerr.Is(err, CodeAccessStoreUnavailable) {
		t.Fatalf("expected store unavailable error, got %v", err)
	}
}

// TestServiceRejectsInvalidInput verifies empty access checks fail before
// hitting storage.
func TestServiceRejectsInvalidInput(t *testing.T) {
	store := &accessTestStore{agentOK: true}
	service, err := New(store)
	if err != nil {
		t.Fatal(err)
	}

	err = service.EnsureWorkspaceResourceVisible(context.Background(), "usr-current", " ", "/home/agent/workspace/app.log")
	if !bizerr.Is(err, CodeAccessInvalidInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
	if store.agentChecks != 0 || store.chatChecks != 0 || store.terminalChecks != 0 {
		t.Fatalf("expected invalid input to avoid store access, got agent=%d chat=%d terminal=%d checks", store.agentChecks, store.chatChecks, store.terminalChecks)
	}
}

type accessTestStore struct {
	agentOK        bool
	chatOK         bool
	terminalOK     bool
	err            error
	agentChecks    int
	chatChecks     int
	terminalChecks int
}

func (s *accessTestStore) UserOwnsAgent(context.Context, string, string) (bool, error) {
	s.agentChecks++
	return s.agentOK, s.err
}

func (s *accessTestStore) UserOwnsChatSession(context.Context, string, string, string) (bool, error) {
	s.chatChecks++
	return s.chatOK, s.err
}

func (s *accessTestStore) UserOwnsTerminalSession(context.Context, string, string, string) (bool, error) {
	s.terminalChecks++
	return s.terminalOK, s.err
}

func assertAccessErrorHasNoLeak(t *testing.T, err error) {
	t.Helper()
	text := strings.ToLower(err.Error())
	for _, forbidden := range []string{
		"secret agent",
		"agt-secret",
		"chat-secret",
		"term-secret",
		"svc-secret",
		"private.log",
		"running",
		"message_count",
		"/home/agent/workspace",
	} {
		if strings.Contains(text, strings.ToLower(forbidden)) {
			t.Fatalf("access error leaked %q: %v", forbidden, err)
		}
	}
}
