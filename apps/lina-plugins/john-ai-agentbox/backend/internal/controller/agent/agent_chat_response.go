// This file converts AgentBox Chat service projections into versioned API DTOs
// without exposing service internals through controller method signatures.

package agent

import (
	v1 "john-ai-agentbox/backend/api/agent/v1"
	chatsvc "john-ai-agentbox/backend/internal/service/chat"
)

func toChatSessionResponse(item chatsvc.SessionInfo) v1.ChatSessionInfo {
	return v1.ChatSessionInfo{
		ID:                 item.ID,
		AgentID:            item.AgentID,
		Title:              item.Title,
		Status:             item.Status,
		ToolType:           item.ToolType,
		ToolSessionID:      item.ToolSessionID,
		RuntimeState:       item.RuntimeState,
		LastError:          item.LastError,
		MessageCount:       item.MessageCount,
		LastMessagePreview: item.LastMessagePreview,
		CreatedAt:          item.CreatedAt,
		UpdatedAt:          item.UpdatedAt,
		LastActiveAt:       item.LastActiveAt,
	}
}

func toChatSessionListResponse(items []chatsvc.SessionInfo) []v1.ChatSessionInfo {
	out := make([]v1.ChatSessionInfo, 0, len(items))
	for _, item := range items {
		out = append(out, toChatSessionResponse(item))
	}
	return out
}

func toChatMessageResponse(item chatsvc.MessageInfo) v1.ChatMessageInfo {
	return v1.ChatMessageInfo{
		ID:        item.ID,
		SessionID: item.SessionID,
		Sequence:  item.Sequence,
		Role:      item.Role,
		Content:   item.Content,
		Status:    item.Status,
		Metadata:  item.Metadata,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

func toChatMessageListResponse(items []chatsvc.MessageInfo) []v1.ChatMessageInfo {
	out := make([]v1.ChatMessageInfo, 0, len(items))
	for _, item := range items {
		out = append(out, toChatMessageResponse(item))
	}
	return out
}

func toChatMessagesResponse(item chatsvc.MessagesOutput) v1.ChatMessagesResponse {
	return v1.ChatMessagesResponse{
		Session:  toChatSessionResponse(item.Session),
		Messages: toChatMessageListResponse(item.Messages),
	}
}

func toChatInteractionResponse(item chatsvc.InteractionInfo) v1.ChatInteractionInfo {
	return v1.ChatInteractionInfo{
		ID:                 item.ID,
		AgentID:            item.AgentID,
		SessionID:          item.SessionID,
		AssistantMessageID: item.AssistantMessageID,
		ToolType:           item.ToolType,
		ToolInteractionID:  item.ToolInteractionID,
		Type:               item.Type,
		Status:             item.Status,
		Title:              item.Title,
		Body:               item.Body,
		RiskLevel:          item.RiskLevel,
		Payload:            item.Payload,
		Response:           item.Response,
		ResponseMode:       item.ResponseMode,
		ResponseScope:      item.ResponseScope,
		ExpiresAt:          item.ExpiresAt,
		ResolvedAt:         item.ResolvedAt,
		CreatedAt:          item.CreatedAt,
		UpdatedAt:          item.UpdatedAt,
	}
}

func toChatInteractionListResponse(items []chatsvc.InteractionInfo) []v1.ChatInteractionInfo {
	out := make([]v1.ChatInteractionInfo, 0, len(items))
	for _, item := range items {
		out = append(out, toChatInteractionResponse(item))
	}
	return out
}

func toChatRecoverResponse(item *chatsvc.RecoverOutput) v1.ChatRecoverResponse {
	if item == nil {
		return v1.ChatRecoverResponse{}
	}
	var response v1.ChatRecoverResponse
	if item.Session != nil {
		session := toChatSessionResponse(*item.Session)
		response.Session = &session
	}
	if item.Message != nil {
		message := toChatMessageResponse(*item.Message)
		response.Message = &message
	}
	return response
}
