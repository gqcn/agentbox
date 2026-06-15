// This file converts AgentBox terminal service projections into versioned API
// DTOs without exposing service internals through controller method signatures.

package agent

import (
	v1 "john-ai-agentbox/backend/api/agent/v1"
	terminalsvc "john-ai-agentbox/backend/internal/service/terminal"
)

func toTerminalSessionResponse(item terminalsvc.SessionInfo) v1.TerminalSessionInfo {
	return v1.TerminalSessionInfo{
		ID:                 item.ID,
		UserID:             item.UserID,
		AgentID:            item.AgentID,
		TerminalID:         item.TerminalID,
		BackendType:        item.BackendType,
		BackendSessionName: item.BackendSessionName,
		WorkingDir:         item.WorkingDir,
		Shell:              item.Shell,
		Status:             item.Status,
		LastError:          item.LastError,
		ClosedAt:           item.ClosedAt,
		CreatedAt:          item.CreatedAt,
		UpdatedAt:          item.UpdatedAt,
	}
}

func toTerminalSessionListResponse(items []terminalsvc.SessionInfo) []v1.TerminalSessionInfo {
	out := make([]v1.TerminalSessionInfo, 0, len(items))
	for _, item := range items {
		out = append(out, toTerminalSessionResponse(item))
	}
	return out
}
