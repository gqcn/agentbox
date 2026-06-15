// This file implements version-one Agent Chat request handlers. Every handler
// derives the AgentBox user ID from authentication context before delegating to
// Chat service methods that enforce Agent/session ownership boundaries.

package agent

import (
	"context"

	v1 "john-ai-agentbox/backend/api/agent/v1"
	"john-ai-agentbox/backend/internal/service/authctx"
	chatsvc "john-ai-agentbox/backend/internal/service/chat"
)

// ChatSessions returns the current user's Chat sessions for one Agent.
func (c *ControllerV1) ChatSessions(ctx context.Context, req *v1.ChatSessionsReq) (res *v1.ChatSessionsRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	items, err := c.chatSvc.ListSessions(ctx, userID, req.ID)
	if err != nil {
		return nil, err
	}
	out := toChatSessionListResponse(items)
	return (*v1.ChatSessionsRes)(&out), nil
}

// CreateChatSession creates one Chat session owned by the current user.
func (c *ControllerV1) CreateChatSession(ctx context.Context, req *v1.CreateChatSessionReq) (res *v1.CreateChatSessionRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.chatSvc.CreateSession(ctx, userID, req.ID)
	if err != nil {
		return nil, err
	}
	out := toChatSessionResponse(*item)
	return (*v1.CreateChatSessionRes)(&out), nil
}

// ChatSession returns one Chat session owned by the current user.
func (c *ControllerV1) ChatSession(ctx context.Context, req *v1.ChatSessionReq) (res *v1.ChatSessionRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.chatSvc.GetSession(ctx, userID, req.ID, req.SessionID)
	if err != nil {
		return nil, err
	}
	out := toChatSessionResponse(*item)
	return (*v1.ChatSessionRes)(&out), nil
}

// UpdateChatSession updates one Chat session title.
func (c *ControllerV1) UpdateChatSession(ctx context.Context, req *v1.UpdateChatSessionReq) (res *v1.UpdateChatSessionRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.chatSvc.UpdateSessionTitle(ctx, userID, req.ID, req.SessionID, req.Title)
	if err != nil {
		return nil, err
	}
	out := toChatSessionResponse(*item)
	return (*v1.UpdateChatSessionRes)(&out), nil
}

// DeleteChatSession deletes one visible Chat session.
func (c *ControllerV1) DeleteChatSession(ctx context.Context, req *v1.DeleteChatSessionReq) (res *v1.DeleteChatSessionRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.chatSvc.DeleteSession(ctx, userID, req.ID, req.SessionID); err != nil {
		return nil, err
	}
	return &v1.DeleteChatSessionRes{Deleted: true}, nil
}

// ChatMessages returns persisted messages for one visible Chat session.
func (c *ControllerV1) ChatMessages(ctx context.Context, req *v1.ChatMessagesReq) (res *v1.ChatMessagesRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.chatSvc.Messages(ctx, userID, req.ID, req.SessionID)
	if err != nil {
		return nil, err
	}
	out := toChatMessagesResponse(*item)
	return (*v1.ChatMessagesRes)(&out), nil
}

// ChatInteractions returns persisted interactions for one visible Chat session.
func (c *ControllerV1) ChatInteractions(ctx context.Context, req *v1.ChatInteractionsReq) (res *v1.ChatInteractionsRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	items, err := c.chatSvc.ListInteractions(ctx, userID, req.ID, req.SessionID, chatsvc.InteractionFilter{
		Status: req.Status,
		Type:   req.Type,
	})
	if err != nil {
		return nil, err
	}
	out := toChatInteractionListResponse(items)
	return (*v1.ChatInteractionsRes)(&out), nil
}

// ChatInteraction returns one visible Chat interaction.
func (c *ControllerV1) ChatInteraction(ctx context.Context, req *v1.ChatInteractionReq) (res *v1.ChatInteractionRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.chatSvc.GetInteraction(ctx, userID, req.ID, req.SessionID, req.InteractionID)
	if err != nil {
		return nil, err
	}
	out := toChatInteractionResponse(*item)
	return (*v1.ChatInteractionRes)(&out), nil
}

// UpdateChatInteractionResponse records a response for one pending interaction.
func (c *ControllerV1) UpdateChatInteractionResponse(ctx context.Context, req *v1.UpdateChatInteractionResponseReq) (res *v1.UpdateChatInteractionResponseRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.chatSvc.UpdateInteractionResponse(ctx, userID, req.ID, req.SessionID, req.InteractionID, chatsvc.InteractionResponseInput{
		Response:      req.Response,
		ResponseMode:  req.ResponseMode,
		ResponseScope: req.ResponseScope,
	})
	if err != nil {
		return nil, err
	}
	out := toChatInteractionResponse(*item)
	return (*v1.UpdateChatInteractionResponseRes)(&out), nil
}

// UpdateChatInteractionStatus updates a pending interaction to a terminal status.
func (c *ControllerV1) UpdateChatInteractionStatus(ctx context.Context, req *v1.UpdateChatInteractionStatusReq) (res *v1.UpdateChatInteractionStatusRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.chatSvc.UpdateInteractionStatus(ctx, userID, req.ID, req.SessionID, req.InteractionID, chatsvc.InteractionStatusInput{
		Status: req.Status,
	})
	if err != nil {
		return nil, err
	}
	out := toChatInteractionResponse(*item)
	return (*v1.UpdateChatInteractionStatusRes)(&out), nil
}

// RecoverChat validates session visibility and starts recovery when runtime support is available.
func (c *ControllerV1) RecoverChat(ctx context.Context, req *v1.RecoverChatReq) (res *v1.RecoverChatRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.chatSvc.Recover(ctx, userID, req.ID, req.SessionID)
	if err != nil {
		return nil, err
	}
	out := toChatRecoverResponse(item)
	return (*v1.RecoverChatRes)(&out), nil
}
