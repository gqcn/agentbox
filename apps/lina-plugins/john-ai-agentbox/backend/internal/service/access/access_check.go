// This file implements AgentBox access checks with invisible-resource semantics.
// Missing resources and resources owned by another AgentBox user are returned
// as the same structured error to avoid existence and metadata leaks.

package access

import (
	"context"
	"strings"

	"lina-core/pkg/bizerr"
)

// EnsureAgentVisible verifies that the current AgentBox user can access the Agent.
func (s *serviceImpl) EnsureAgentVisible(ctx context.Context, userID string, agentID string) error {
	userID, agentID = normalizeOwnerAndAgent(userID, agentID)
	if userID == "" || agentID == "" {
		return bizerr.NewCode(CodeAccessInvalidInput)
	}
	ok, err := s.store.UserOwnsAgent(ctx, userID, agentID)
	return accessDecision(ok, err)
}

// EnsureChatSessionVisible verifies that the current AgentBox user can access the Agent Chat session.
func (s *serviceImpl) EnsureChatSessionVisible(ctx context.Context, userID string, agentID string, sessionID string) error {
	userID, agentID = normalizeOwnerAndAgent(userID, agentID)
	sessionID = strings.TrimSpace(sessionID)
	if userID == "" || agentID == "" || sessionID == "" {
		return bizerr.NewCode(CodeAccessInvalidInput)
	}
	ok, err := s.store.UserOwnsChatSession(ctx, userID, agentID, sessionID)
	return accessDecision(ok, err)
}

// EnsureTerminalSessionVisible verifies that the current AgentBox user can access the persisted terminal session.
func (s *serviceImpl) EnsureTerminalSessionVisible(ctx context.Context, userID string, agentID string, terminalID string) error {
	userID, agentID = normalizeOwnerAndAgent(userID, agentID)
	terminalID = strings.TrimSpace(terminalID)
	if userID == "" || agentID == "" || terminalID == "" {
		return bizerr.NewCode(CodeAccessInvalidInput)
	}
	ok, err := s.store.UserOwnsTerminalSession(ctx, userID, agentID, terminalID)
	return accessDecision(ok, err)
}

// EnsureWorkspaceResourceVisible verifies Agent visibility for workspace resources and downloads.
func (s *serviceImpl) EnsureWorkspaceResourceVisible(ctx context.Context, userID string, agentID string, path string) error {
	userID, agentID = normalizeOwnerAndAgent(userID, agentID)
	path = strings.TrimSpace(path)
	if userID == "" || agentID == "" || path == "" {
		return bizerr.NewCode(CodeAccessInvalidInput)
	}
	ok, err := s.store.UserOwnsAgent(ctx, userID, agentID)
	return accessDecision(ok, err)
}

// EnsureServiceProxyVisible verifies Agent visibility for runtime service proxy and tunnel requests.
func (s *serviceImpl) EnsureServiceProxyVisible(ctx context.Context, userID string, agentID string, serviceID string) error {
	userID, agentID = normalizeOwnerAndAgent(userID, agentID)
	serviceID = strings.TrimSpace(serviceID)
	if userID == "" || agentID == "" || serviceID == "" {
		return bizerr.NewCode(CodeAccessInvalidInput)
	}
	ok, err := s.store.UserOwnsAgent(ctx, userID, agentID)
	return accessDecision(ok, err)
}

func normalizeOwnerAndAgent(userID string, agentID string) (string, string) {
	return strings.TrimSpace(userID), strings.TrimSpace(agentID)
}

func accessDecision(ok bool, err error) error {
	if err != nil {
		return bizerr.WrapCode(err, CodeAccessStoreUnavailable)
	}
	if !ok {
		return bizerr.NewCode(CodeAccessResourceUnavailable)
	}
	return nil
}
