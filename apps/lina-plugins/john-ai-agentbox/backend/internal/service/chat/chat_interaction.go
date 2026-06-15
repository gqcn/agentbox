// This file implements Chat interaction reads and state transitions. Each
// interaction query is scoped by Agent ID and session ID after ownership
// validation, and updates only affect pending rows to prevent duplicate user
// responses.

package chat

import (
	"context"
	"strings"
	"time"

	"lina-core/pkg/bizerr"

	"john-ai-agentbox/backend/internal/dao"
	"john-ai-agentbox/backend/internal/model/do"
	"john-ai-agentbox/backend/internal/model/entity"
)

// ListInteractions returns visible interactions for one Chat session.
func (s *serviceImpl) ListInteractions(ctx context.Context, userID string, agentID string, sessionID string, filter InteractionFilter) ([]InteractionInfo, error) {
	userID, agentID, sessionID = normalizeOwnerAgentSession(userID, agentID, sessionID)
	if err := s.accessSvc.EnsureChatSessionVisible(ctx, userID, agentID, sessionID); err != nil {
		return nil, err
	}
	query := dao.AgentChatInteractions.Ctx(ctx).
		Where(do.AgentChatInteractions{
			AgentId:   agentID,
			SessionId: sessionID,
		})
	status := strings.TrimSpace(filter.Status)
	if status != "" {
		query = query.Where(do.AgentChatInteractions{Status: normalizeInteractionStatus(status)})
	}
	interactionType := strings.TrimSpace(filter.Type)
	if interactionType != "" {
		query = query.Where(do.AgentChatInteractions{InteractionType: normalizeInteractionType(interactionType)})
	}
	cols := dao.AgentChatInteractions.Columns()
	var rows []*entity.AgentChatInteractions
	err := query.OrderAsc(cols.CreatedAt).OrderAsc(cols.Id).Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeChatStoreUnavailable)
	}
	items := make([]InteractionInfo, 0, len(rows))
	for _, row := range rows {
		items = append(items, interactionFromEntity(row))
	}
	return items, nil
}

// GetInteraction returns one visible interaction for a Chat session.
func (s *serviceImpl) GetInteraction(ctx context.Context, userID string, agentID string, sessionID string, interactionID string) (*InteractionInfo, error) {
	userID, agentID, sessionID = normalizeOwnerAgentSession(userID, agentID, sessionID)
	interactionID = strings.TrimSpace(interactionID)
	if interactionID == "" {
		return nil, bizerr.NewCode(CodeChatInvalidInput)
	}
	if err := s.accessSvc.EnsureChatSessionVisible(ctx, userID, agentID, sessionID); err != nil {
		return nil, err
	}
	return getInteraction(ctx, agentID, sessionID, interactionID)
}

// UpdateInteractionResponse stores a user response for one pending interaction.
func (s *serviceImpl) UpdateInteractionResponse(ctx context.Context, userID string, agentID string, sessionID string, interactionID string, input InteractionResponseInput) (*InteractionInfo, error) {
	userID, agentID, sessionID = normalizeOwnerAgentSession(userID, agentID, sessionID)
	interactionID = strings.TrimSpace(interactionID)
	if interactionID == "" {
		return nil, bizerr.NewCode(CodeChatInvalidInput)
	}
	if err := s.accessSvc.EnsureChatSessionVisible(ctx, userID, agentID, sessionID); err != nil {
		return nil, err
	}
	response, ok := defaultJSON(input.Response)
	if !ok {
		return nil, bizerr.NewCode(CodeChatInvalidInput)
	}
	responseMode := normalizeInteractionResponseMode(input.ResponseMode)
	now := time.Now()
	result, err := dao.AgentChatInteractions.Ctx(ctx).
		Where(do.AgentChatInteractions{
			Id:        interactionID,
			AgentId:   agentID,
			SessionId: sessionID,
			Status:    InteractionStatusPending,
		}).
		Data(do.AgentChatInteractions{
			Status:        terminalStatusFromResponseMode(responseMode),
			ResponseJson:  response,
			ResponseMode:  responseMode,
			ResponseScope: normalizeInteractionResponseScope(input.ResponseScope),
			ResolvedAt:    &now,
		}).
		Update()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeChatStoreUnavailable)
	}
	if affected, affectedErr := result.RowsAffected(); affectedErr == nil && affected == 0 {
		return nil, bizerr.NewCode(CodeChatStateConflict)
	}
	return getInteraction(ctx, agentID, sessionID, interactionID)
}

// UpdateInteractionStatus updates a pending interaction to a terminal status.
func (s *serviceImpl) UpdateInteractionStatus(ctx context.Context, userID string, agentID string, sessionID string, interactionID string, input InteractionStatusInput) (*InteractionInfo, error) {
	userID, agentID, sessionID = normalizeOwnerAgentSession(userID, agentID, sessionID)
	interactionID = strings.TrimSpace(interactionID)
	status := normalizeInteractionControlStatus(input.Status)
	if interactionID == "" || status == "" {
		return nil, bizerr.NewCode(CodeChatInvalidInput)
	}
	if err := s.accessSvc.EnsureChatSessionVisible(ctx, userID, agentID, sessionID); err != nil {
		return nil, err
	}
	now := time.Now()
	result, err := dao.AgentChatInteractions.Ctx(ctx).
		Where(do.AgentChatInteractions{
			Id:        interactionID,
			AgentId:   agentID,
			SessionId: sessionID,
			Status:    InteractionStatusPending,
		}).
		Data(do.AgentChatInteractions{
			Status:     status,
			ResolvedAt: &now,
		}).
		Update()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeChatStoreUnavailable)
	}
	if affected, affectedErr := result.RowsAffected(); affectedErr == nil && affected == 0 {
		return nil, bizerr.NewCode(CodeChatStateConflict)
	}
	return getInteraction(ctx, agentID, sessionID, interactionID)
}

func getInteraction(ctx context.Context, agentID string, sessionID string, interactionID string) (*InteractionInfo, error) {
	var row *entity.AgentChatInteractions
	err := dao.AgentChatInteractions.Ctx(ctx).
		Where(do.AgentChatInteractions{
			Id:        strings.TrimSpace(interactionID),
			AgentId:   strings.TrimSpace(agentID),
			SessionId: strings.TrimSpace(sessionID),
		}).
		Scan(&row)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeChatStoreUnavailable)
	}
	if row == nil {
		return nil, bizerr.NewCode(CodeChatNotFound)
	}
	item := interactionFromEntity(row)
	return &item, nil
}

func interactionFromEntity(row *entity.AgentChatInteractions) InteractionInfo {
	if row == nil {
		return InteractionInfo{}
	}
	return InteractionInfo{
		ID:                 row.Id,
		AgentID:            row.AgentId,
		SessionID:          row.SessionId,
		AssistantMessageID: row.AssistantMessageId,
		ToolType:           row.ToolType,
		ToolInteractionID:  row.ToolInteractionId,
		Type:               normalizeInteractionType(row.InteractionType),
		Status:             normalizeInteractionStatus(row.Status),
		Title:              row.Title,
		Body:               row.Body,
		RiskLevel:          normalizeInteractionRisk(row.RiskLevel),
		Payload:            row.PayloadJson,
		Response:           row.ResponseJson,
		ResponseMode:       normalizeInteractionResponseMode(row.ResponseMode),
		ResponseScope:      normalizeInteractionResponseScope(row.ResponseScope),
		ExpiresAt:          unixMilliPtrFromTimePtr(row.ExpiresAt),
		ResolvedAt:         unixMilliPtrFromTimePtr(row.ResolvedAt),
		CreatedAt:          unixMilliFromTimePtr(row.CreatedAt),
		UpdatedAt:          unixMilliFromTimePtr(row.UpdatedAt),
	}
}

func normalizeInteractionRisk(value string) string {
	switch strings.TrimSpace(value) {
	case InteractionRiskLow, InteractionRiskMedium, InteractionRiskHigh, InteractionRiskCritical:
		return strings.TrimSpace(value)
	default:
		return InteractionRiskLow
	}
}
