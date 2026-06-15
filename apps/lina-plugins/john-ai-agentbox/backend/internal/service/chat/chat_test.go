// This file verifies Chat service construction and access-denied behavior. The
// tests use a fake access service so invisible Agent/session resources are
// rejected before any DAO query can expose cross-user metadata.

package chat

import (
	"context"
	"testing"

	"lina-core/pkg/bizerr"

	accesssvc "john-ai-agentbox/backend/internal/service/access"
)

// TestNewRequiresAccessService verifies missing runtime dependencies fail at construction.
func TestNewRequiresAccessService(t *testing.T) {
	if _, err := New(nil); err == nil {
		t.Fatal("expected constructor to reject nil access service")
	}
}

// TestListSessionsPropagatesInvisibleAgent verifies invisible Agents are rejected before DAO reads.
func TestListSessionsPropagatesInvisibleAgent(t *testing.T) {
	service, err := New(&fakeAccessService{err: bizerr.NewCode(accesssvc.CodeAccessResourceUnavailable)})
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.ListSessions(context.Background(), "usr-owner", "agt-other")
	if !bizerr.Is(err, accesssvc.CodeAccessResourceUnavailable) {
		t.Fatalf("expected invisible resource error, got %v", err)
	}
}

// TestMessagesPropagatesInvisibleChatSession verifies invisible sessions do not leak metadata.
func TestMessagesPropagatesInvisibleChatSession(t *testing.T) {
	service, err := New(&fakeAccessService{err: bizerr.NewCode(accesssvc.CodeAccessResourceUnavailable)})
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.Messages(context.Background(), "usr-owner", "agt-test", "chat-other")
	if !bizerr.Is(err, accesssvc.CodeAccessResourceUnavailable) {
		t.Fatalf("expected invisible session error, got %v", err)
	}
}

// TestRecoverChecksSessionVisibilityBeforeRuntime verifies runtime recovery does not bypass ownership.
func TestRecoverChecksSessionVisibilityBeforeRuntime(t *testing.T) {
	service, err := New(&fakeAccessService{err: bizerr.NewCode(accesssvc.CodeAccessResourceUnavailable)})
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.Recover(context.Background(), "usr-owner", "agt-test", "chat-other")
	if !bizerr.Is(err, accesssvc.CodeAccessResourceUnavailable) {
		t.Fatalf("expected invisible session error, got %v", err)
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
