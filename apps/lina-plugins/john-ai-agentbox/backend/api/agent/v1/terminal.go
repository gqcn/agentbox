// This file defines AgentBox terminal-session API DTOs. The routes are plugin
// relative and are published under /x/john-ai-agentbox/api/v1 by source-plugin
// route registration.

package v1

import "github.com/gogf/gf/v2/frame/g"

// TerminalSessionsReq lists persisted terminal sessions for one Agent.
type TerminalSessionsReq struct {
	g.Meta `path:"/agents/{id}/terminal/sessions" method:"get" tags:"AgentBox Terminal" summary:"List AgentBox terminal sessions" dc:"List persisted terminal sessions for one coding agent owned by the authenticated AgentBox user."`
	ID     string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	Status string `json:"status" dc:"Optional terminal session status filter: active=attachable, closed=closed, error=recovery failed; empty returns all statuses" eg:"active"`
}

// TerminalSessionsRes returns terminal sessions.
type TerminalSessionsRes = []TerminalSessionInfo

// CreateTerminalSessionReq creates or rebuilds one terminal session metadata row.
type CreateTerminalSessionReq struct {
	g.Meta     `path:"/agents/{id}/terminal/sessions" method:"post" tags:"AgentBox Terminal" summary:"Create AgentBox terminal session" dc:"Create or rebuild the persisted terminal session metadata for one authenticated-user-owned Agent before a Shell WebSocket attaches to the runtime backend."`
	ID         string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	TerminalID string `json:"terminalId" v:"required" dc:"Browser terminal pane ID; must be scoped by the current AgentBox user and Agent" eg:"shell-terminal-agt-123-1700000000000-1"`
	WorkingDir string `json:"workingDir" dc:"Container working directory requested for this terminal session; empty uses the Agent workspace root" eg:"/home/agent/workspace"`
	Shell      string `json:"shell" dc:"Shell executable requested for this terminal session; empty uses the image default shell" eg:"/bin/bash"`
	Rebuild    bool   `json:"rebuild" dc:"Whether to reset a closed or errored terminal session using the same browser terminal ID" eg:"false"`
}

// CreateTerminalSessionRes returns the created terminal session metadata.
type CreateTerminalSessionRes = TerminalSessionInfo

// TerminalSessionReq gets one terminal session.
type TerminalSessionReq struct {
	g.Meta     `path:"/agents/{id}/terminal/sessions/{terminalId}" method:"get" tags:"AgentBox Terminal" summary:"Get AgentBox terminal session" dc:"Get one persisted terminal session for one AgentBox-user-owned coding agent without exposing other users' terminal metadata."`
	ID         string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	TerminalID string `json:"terminalId" v:"required" dc:"Browser terminal pane ID" eg:"shell-terminal-agt-123-1700000000000-1"`
}

// TerminalSessionRes returns one terminal session.
type TerminalSessionRes = TerminalSessionInfo

// CloseTerminalSessionReq closes one terminal session.
type CloseTerminalSessionReq struct {
	g.Meta     `path:"/agents/{id}/terminal/sessions/{terminalId}" method:"delete" tags:"AgentBox Terminal" summary:"Close AgentBox terminal session" dc:"Mark one authenticated-user-owned terminal session closed so reconnects require an explicit rebuild."`
	ID         string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	TerminalID string `json:"terminalId" v:"required" dc:"Browser terminal pane ID" eg:"shell-terminal-agt-123-1700000000000-1"`
}

// CloseTerminalSessionRes reports terminal close state.
type CloseTerminalSessionRes struct {
	Closed bool `json:"closed" dc:"Whether the terminal session is now closed" eg:"true"`
}

// TerminalSessionInfo is the public persisted terminal-session projection.
type TerminalSessionInfo struct {
	ID                 string `json:"id" dc:"Terminal session ID" eg:"term-1234567890abcdef1234567890abcdef"`
	UserID             string `json:"userId,omitempty" dc:"AgentBox user ID that owns this terminal session; omitted when not needed by the client" eg:"usr-admin"`
	AgentID            string `json:"agentId" dc:"Agent ID that owns this terminal session" eg:"agt-1234567890abcdef"`
	TerminalID         string `json:"terminalId" dc:"Browser terminal pane ID" eg:"shell-terminal-agt-123-1700000000000-1"`
	BackendType        string `json:"backendType" dc:"Terminal backend type: tmux=tmux-backed persistent session" eg:"tmux"`
	BackendSessionName string `json:"backendSessionName" dc:"Server-generated backend session name; never derived from raw terminal input" eg:"abx-abc123def4567890"`
	WorkingDir         string `json:"workingDir" dc:"Container working directory used when the terminal session was created" eg:"/home/agent/workspace"`
	Shell              string `json:"shell" dc:"Default shell used by this terminal session" eg:"/bin/bash"`
	Status             string `json:"status" dc:"Terminal session status: active=attachable, closed=closed, error=recovery failed" eg:"active"`
	LastError          string `json:"lastError,omitempty" dc:"Latest backend recovery or attach error, omitted when healthy" eg:"tmux session not found"`
	ClosedAt           *int64 `json:"closedAt,omitempty" dc:"Close time as Unix timestamp in milliseconds; omitted while active" eg:"1704067500000"`
	CreatedAt          int64  `json:"createdAt" dc:"Creation time as Unix timestamp in milliseconds" eg:"1704067200000"`
	UpdatedAt          int64  `json:"updatedAt" dc:"Last update time as Unix timestamp in milliseconds" eg:"1704067300000"`
}
