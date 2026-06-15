// This file verifies terminal metadata boundaries that do not require a real
// Shell WebSocket runtime. Ownership must be checked before session state is
// exposed or reset.

package terminal

import (
	"context"
	"strings"
	"testing"

	"lina-core/pkg/bizerr"
)

// TestBackendSessionNameDoesNotExposeTerminalID verifies raw browser terminal
// IDs are never used as backend tmux session names.
func TestBackendSessionNameDoesNotExposeTerminalID(t *testing.T) {
	name := backendSessionName("usr-current", "agt-owned", "terminal; rm -rf /")
	if !strings.HasPrefix(name, "abx-") {
		t.Fatalf("backend session name = %q, want abx prefix", name)
	}
	if strings.Contains(name, "terminal") || strings.Contains(name, ";") || strings.Contains(name, "/") {
		t.Fatalf("backend session name contains raw input: %q", name)
	}
	if again := backendSessionName("usr-current", "agt-owned", "terminal; rm -rf /"); again != name {
		t.Fatalf("backend session name is not deterministic: %q != %q", again, name)
	}
}

// TestListSessionsChecksAgentVisibility verifies list calls prove Agent
// ownership before terminal metadata can be returned.
func TestListSessionsChecksAgentVisibility(t *testing.T) {
	access := &terminalAccessStub{err: bizerr.NewCode(CodeTerminalNotFound)}
	service, err := New(access)
	if err != nil {
		t.Fatalf("new terminal service: %v", err)
	}
	_, err = service.ListSessions(context.Background(), "usr-current", "agt-other", SessionFilter{})
	if err == nil {
		t.Fatal("expected ownership error")
	}
	if access.agentChecks != 1 || access.terminalChecks != 0 {
		t.Fatalf("unexpected access checks: agent=%d terminal=%d", access.agentChecks, access.terminalChecks)
	}
}

// TestGetSessionChecksTerminalVisibility verifies direct terminal lookup uses
// terminal-specific ownership checks before loading state.
func TestGetSessionChecksTerminalVisibility(t *testing.T) {
	access := &terminalAccessStub{err: bizerr.NewCode(CodeTerminalNotFound)}
	service, err := New(access)
	if err != nil {
		t.Fatalf("new terminal service: %v", err)
	}
	_, err = service.GetSession(context.Background(), "usr-current", "agt-other", "term-secret")
	if err == nil {
		t.Fatal("expected ownership error")
	}
	if access.terminalChecks != 1 || access.lastTerminalID != "term-secret" {
		t.Fatalf("unexpected terminal check: count=%d id=%q", access.terminalChecks, access.lastTerminalID)
	}
}

// TestInvalidStatusRejected verifies optional status filters are bounded.
func TestInvalidStatusRejected(t *testing.T) {
	access := &terminalAccessStub{}
	service, err := New(access)
	if err != nil {
		t.Fatalf("new terminal service: %v", err)
	}
	_, err = service.ListSessions(context.Background(), "usr-current", "agt-owned", SessionFilter{Status: "running"})
	if !bizerr.Is(err, CodeTerminalInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

type terminalAccessStub struct {
	err            error
	agentChecks    int
	terminalChecks int
	lastTerminalID string
}

func (s *terminalAccessStub) EnsureAgentVisible(context.Context, string, string) error {
	s.agentChecks++
	return s.err
}

func (s *terminalAccessStub) EnsureChatSessionVisible(context.Context, string, string, string) error {
	return s.err
}

func (s *terminalAccessStub) EnsureTerminalSessionVisible(_ context.Context, _ string, _ string, terminalID string) error {
	s.terminalChecks++
	s.lastTerminalID = terminalID
	return s.err
}

func (s *terminalAccessStub) EnsureWorkspaceResourceVisible(context.Context, string, string, string) error {
	return s.err
}

func (s *terminalAccessStub) EnsureServiceProxyVisible(context.Context, string, string, string) error {
	return s.err
}
