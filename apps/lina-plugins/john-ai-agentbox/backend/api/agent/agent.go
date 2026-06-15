// This file defines the public agent API contract surface for the AgentBox
// plugin. DTOs live under versioned subpackages while this package exposes the
// GoFrame controller interface used by route registration.

package agent

import (
	"context"

	v1 "john-ai-agentbox/backend/api/agent/v1"
)

// IAgentV1 defines AgentBox coding-agent HTTP handlers.
type IAgentV1 interface {
	// List returns the authenticated user's non-deleted coding agents.
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
	// Create creates one coding agent owned by the authenticated user.
	Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error)
	// Detail returns one authenticated-user-owned coding agent.
	Detail(ctx context.Context, req *v1.DetailReq) (res *v1.DetailRes, err error)
	// Update updates one authenticated-user-owned coding agent.
	Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error)
	// ChangeImage switches one authenticated-user-owned coding agent image.
	ChangeImage(ctx context.Context, req *v1.ChangeImageReq) (res *v1.ChangeImageRes, err error)
	// Start starts one authenticated-user-owned coding agent runtime when available.
	Start(ctx context.Context, req *v1.StartReq) (res *v1.StartRes, err error)
	// Stop stops one authenticated-user-owned coding agent runtime when available.
	Stop(ctx context.Context, req *v1.StopReq) (res *v1.StopRes, err error)
	// Logs reads one authenticated-user-owned coding agent runtime log when available.
	Logs(ctx context.Context, req *v1.LogsReq) (res *v1.LogsRes, err error)
	// Delete soft-deletes one authenticated-user-owned coding agent.
	Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error)
	// ChatSessions lists Chat sessions for one authenticated-user-owned agent.
	ChatSessions(ctx context.Context, req *v1.ChatSessionsReq) (res *v1.ChatSessionsRes, err error)
	// CreateChatSession creates one empty Chat session for an authenticated-user-owned agent.
	CreateChatSession(ctx context.Context, req *v1.CreateChatSessionReq) (res *v1.CreateChatSessionRes, err error)
	// ChatSession gets one Chat session for an authenticated-user-owned agent.
	ChatSession(ctx context.Context, req *v1.ChatSessionReq) (res *v1.ChatSessionRes, err error)
	// UpdateChatSession updates one Chat session title.
	UpdateChatSession(ctx context.Context, req *v1.UpdateChatSessionReq) (res *v1.UpdateChatSessionRes, err error)
	// DeleteChatSession deletes one visible Chat session.
	DeleteChatSession(ctx context.Context, req *v1.DeleteChatSessionReq) (res *v1.DeleteChatSessionRes, err error)
	// ChatMessages returns message history for one visible Chat session.
	ChatMessages(ctx context.Context, req *v1.ChatMessagesReq) (res *v1.ChatMessagesRes, err error)
	// ChatInteractions lists interactions for one visible Chat session.
	ChatInteractions(ctx context.Context, req *v1.ChatInteractionsReq) (res *v1.ChatInteractionsRes, err error)
	// ChatInteraction gets one visible Chat interaction.
	ChatInteraction(ctx context.Context, req *v1.ChatInteractionReq) (res *v1.ChatInteractionRes, err error)
	// UpdateChatInteractionResponse records a response for one pending interaction.
	UpdateChatInteractionResponse(ctx context.Context, req *v1.UpdateChatInteractionResponseReq) (res *v1.UpdateChatInteractionResponseRes, err error)
	// UpdateChatInteractionStatus updates one pending interaction to a terminal status.
	UpdateChatInteractionStatus(ctx context.Context, req *v1.UpdateChatInteractionStatusReq) (res *v1.UpdateChatInteractionStatusRes, err error)
	// RecoverChat validates and starts Chat recovery when runtime support is available.
	RecoverChat(ctx context.Context, req *v1.RecoverChatReq) (res *v1.RecoverChatRes, err error)
	// TerminalSessions lists persisted terminal sessions for one authenticated-user-owned Agent.
	TerminalSessions(ctx context.Context, req *v1.TerminalSessionsReq) (res *v1.TerminalSessionsRes, err error)
	// CreateTerminalSession creates or rebuilds one persisted terminal session metadata row.
	CreateTerminalSession(ctx context.Context, req *v1.CreateTerminalSessionReq) (res *v1.CreateTerminalSessionRes, err error)
	// TerminalSession gets one persisted terminal session for an authenticated-user-owned Agent.
	TerminalSession(ctx context.Context, req *v1.TerminalSessionReq) (res *v1.TerminalSessionRes, err error)
	// CloseTerminalSession closes one persisted terminal session without affecting LinaPro sessions.
	CloseTerminalSession(ctx context.Context, req *v1.CloseTerminalSessionReq) (res *v1.CloseTerminalSessionRes, err error)
}
