// Package access owns AgentBox resource visibility checks. AgentBox uses its
// plugin-local users instead of LinaPro role data permissions, so every runtime
// entry point must prove the target resource belongs to the current AgentBox
// user before returning data, opening sockets, or proxying traffic.
package access

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
)

// Store defines database-backed ownership predicates used by the access service.
type Store interface {
	// UserOwnsAgent reports whether one non-deleted Agent belongs to userID.
	// False means missing or invisible and must be returned as the same caller
	// error to avoid leaking other users' resource existence.
	UserOwnsAgent(ctx context.Context, userID string, agentID string) (bool, error)
	// UserOwnsChatSession reports whether sessionID belongs to agentID and the
	// Agent belongs to userID. False must not expose the session title, status,
	// message count, or any other cross-user metadata.
	UserOwnsChatSession(ctx context.Context, userID string, agentID string, sessionID string) (bool, error)
	// UserOwnsTerminalSession reports whether terminalID belongs to agentID and
	// userID. False must not expose shell, working directory, backend session,
	// status, or other cross-user metadata.
	UserOwnsTerminalSession(ctx context.Context, userID string, agentID string, terminalID string) (bool, error)
}

// Service defines AgentBox resource visibility checks for runtime entry points.
type Service interface {
	// EnsureAgentVisible verifies that the current AgentBox user can access the
	// Agent. This is the prerequisite for workspace, service proxy, and file
	// download operations that are scoped directly by Agent ID.
	EnsureAgentVisible(ctx context.Context, userID string, agentID string) error
	// EnsureChatSessionVisible verifies that the current AgentBox user can access
	// the Agent Chat session before listing messages or opening Chat WebSocket.
	EnsureChatSessionVisible(ctx context.Context, userID string, agentID string, sessionID string) error
	// EnsureTerminalSessionVisible verifies that the current AgentBox user can
	// access the persisted terminal session before reconnecting or closing it.
	EnsureTerminalSessionVisible(ctx context.Context, userID string, agentID string, terminalID string) error
	// EnsureWorkspaceResourceVisible verifies Agent visibility for workspace
	// resources and downloads. Path traversal and file-type checks remain owned
	// by the workspace implementation after this ownership boundary succeeds.
	EnsureWorkspaceResourceVisible(ctx context.Context, userID string, agentID string, path string) error
	// EnsureServiceProxyVisible verifies Agent visibility for runtime service
	// proxy and tunnel requests. Service discovery and bridge key checks remain
	// owned by the runtime gateway after this ownership boundary succeeds.
	EnsureServiceProxyVisible(ctx context.Context, userID string, agentID string, serviceID string) error
}

// serviceImpl implements Service with a Store dependency.
type serviceImpl struct {
	store Store
}

var _ Service = (*serviceImpl)(nil)

// New creates an access service using explicit storage dependency injection.
func New(store Store) (Service, error) {
	if store == nil {
		return nil, gerror.New("agentbox access store is required")
	}
	return &serviceImpl{store: store}, nil
}
