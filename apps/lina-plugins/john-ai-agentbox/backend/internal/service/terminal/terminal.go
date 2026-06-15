// Package terminal owns AgentBox terminal-session metadata. It persists
// browser terminal panes in plugin-owned tables and validates AgentBox user
// ownership before exposing or mutating terminal state; actual Shell WebSocket
// streaming remains owned by the runtime gateway.
package terminal

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	accesssvc "john-ai-agentbox/backend/internal/service/access"
)

const (
	// BackendTypeTmux marks tmux as the current persistent terminal backend.
	BackendTypeTmux = "tmux"
	// SessionStatusActive marks a terminal session that can be attached.
	SessionStatusActive = "active"
	// SessionStatusClosed marks a terminal session explicitly closed by the user.
	SessionStatusClosed = "closed"
	// SessionStatusError marks a terminal session whose backend recovery failed.
	SessionStatusError = "error"
	// DefaultWorkingDir is the default AgentBox runtime workspace directory.
	DefaultWorkingDir = "/home/agent/workspace"
	// DefaultShell is the default terminal shell when no image-specific value is available.
	DefaultShell = "/bin/bash"
)

// SessionInfo is the service-level persisted terminal-session projection.
type SessionInfo struct {
	ID                 string
	UserID             string
	AgentID            string
	TerminalID         string
	BackendType        string
	BackendSessionName string
	WorkingDir         string
	Shell              string
	Status             string
	LastError          string
	ClosedAt           *int64
	CreatedAt          int64
	UpdatedAt          int64
}

// SessionFilter carries optional terminal-session filters.
type SessionFilter struct {
	Status string
}

// SessionInput carries terminal-session creation or rebuild fields.
type SessionInput struct {
	TerminalID string
	WorkingDir string
	Shell      string
	Rebuild    bool
}

// Service defines AgentBox terminal-session metadata behavior. Methods validate
// the current AgentBox user before touching persisted terminal rows so terminal
// IDs cannot disclose other users' shell, working directory, status, or backend
// session names.
type Service interface {
	// ListSessions returns persisted terminal sessions for one visible Agent.
	// Empty status means all statuses; non-empty status is normalized to the
	// supported terminal status constants.
	ListSessions(ctx context.Context, userID string, agentID string, filter SessionFilter) ([]SessionInfo, error)
	// EnsureSession creates or rebuilds terminal metadata for one visible Agent.
	// It does not open a WebSocket or start a runtime process; runtime attach is
	// handled by the gateway after the same AgentBox ownership boundary.
	EnsureSession(ctx context.Context, userID string, agentID string, input SessionInput) (*SessionInfo, error)
	// GetSession returns one visible terminal session by browser terminal ID.
	GetSession(ctx context.Context, userID string, agentID string, terminalID string) (*SessionInfo, error)
	// CloseSession marks one visible terminal session closed. The operation is
	// idempotent for already closed sessions.
	CloseSession(ctx context.Context, userID string, agentID string, terminalID string) error
}

type serviceImpl struct {
	accessSvc accesssvc.Service
}

var _ Service = (*serviceImpl)(nil)

// New creates a terminal service with explicit access dependency injection.
func New(accessSvc accesssvc.Service) (Service, error) {
	if accessSvc == nil {
		return nil, gerror.New("agentbox terminal access service is required")
	}
	return &serviceImpl{accessSvc: accessSvc}, nil
}
