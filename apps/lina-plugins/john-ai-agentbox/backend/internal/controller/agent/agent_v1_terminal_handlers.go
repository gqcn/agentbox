// This file implements version-one Agent terminal metadata handlers. Streaming
// Shell WebSocket handling remains in the gateway; these handlers only manage
// persisted session state after AgentBox user authentication.

package agent

import (
	"context"

	v1 "john-ai-agentbox/backend/api/agent/v1"
	"john-ai-agentbox/backend/internal/service/authctx"
	terminalsvc "john-ai-agentbox/backend/internal/service/terminal"
)

// TerminalSessions returns persisted terminal sessions for one current-user Agent.
func (c *ControllerV1) TerminalSessions(ctx context.Context, req *v1.TerminalSessionsReq) (res *v1.TerminalSessionsRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	items, err := c.terminalSvc.ListSessions(ctx, userID, req.ID, terminalsvc.SessionFilter{Status: req.Status})
	if err != nil {
		return nil, err
	}
	out := toTerminalSessionListResponse(items)
	return (*v1.TerminalSessionsRes)(&out), nil
}

// CreateTerminalSession creates or rebuilds persisted terminal metadata.
func (c *ControllerV1) CreateTerminalSession(ctx context.Context, req *v1.CreateTerminalSessionReq) (res *v1.CreateTerminalSessionRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.terminalSvc.EnsureSession(ctx, userID, req.ID, terminalsvc.SessionInput{
		TerminalID: req.TerminalID,
		WorkingDir: req.WorkingDir,
		Shell:      req.Shell,
		Rebuild:    req.Rebuild,
	})
	if err != nil {
		return nil, err
	}
	out := toTerminalSessionResponse(*item)
	return (*v1.CreateTerminalSessionRes)(&out), nil
}

// TerminalSession returns one persisted terminal session.
func (c *ControllerV1) TerminalSession(ctx context.Context, req *v1.TerminalSessionReq) (res *v1.TerminalSessionRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.terminalSvc.GetSession(ctx, userID, req.ID, req.TerminalID)
	if err != nil {
		return nil, err
	}
	out := toTerminalSessionResponse(*item)
	return (*v1.TerminalSessionRes)(&out), nil
}

// CloseTerminalSession marks one persisted terminal session closed.
func (c *ControllerV1) CloseTerminalSession(ctx context.Context, req *v1.CloseTerminalSessionReq) (res *v1.CloseTerminalSessionRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.terminalSvc.CloseSession(ctx, userID, req.ID, req.TerminalID); err != nil {
		return nil, err
	}
	return &v1.CloseTerminalSessionRes{Closed: true}, nil
}
