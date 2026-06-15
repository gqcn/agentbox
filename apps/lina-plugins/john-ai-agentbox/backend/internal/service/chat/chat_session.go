// This file implements Chat session CRUD and message history reads. Agent and
// session ownership checks happen before DAO queries expose session metadata,
// and list/detail queries are scoped by Agent ID in the database.

package chat

import (
	"context"
	"strings"

	"lina-core/pkg/bizerr"

	"john-ai-agentbox/backend/internal/dao"
	"john-ai-agentbox/backend/internal/model/do"
	"john-ai-agentbox/backend/internal/model/entity"
)

// ListSessions returns all sessions for one current-user-owned Agent.
func (s *serviceImpl) ListSessions(ctx context.Context, userID string, agentID string) ([]SessionInfo, error) {
	userID, agentID = normalizeOwnerAndAgent(userID, agentID)
	if err := s.accessSvc.EnsureAgentVisible(ctx, userID, agentID); err != nil {
		return nil, err
	}
	var rows []*entity.AgentChatSessions
	cols := dao.AgentChatSessions.Columns()
	err := dao.AgentChatSessions.Ctx(ctx).
		Where(do.AgentChatSessions{AgentId: agentID}).
		OrderDesc(cols.LastActiveAt).
		OrderDesc(cols.CreatedAt).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeChatStoreUnavailable)
	}
	items := make([]SessionInfo, 0, len(rows))
	for _, row := range rows {
		items = append(items, sessionFromEntity(row))
	}
	return items, nil
}

// CreateSession creates one empty Chat session for a current-user-owned Agent.
func (s *serviceImpl) CreateSession(ctx context.Context, userID string, agentID string) (*SessionInfo, error) {
	userID, agentID = normalizeOwnerAndAgent(userID, agentID)
	if err := s.accessSvc.EnsureAgentVisible(ctx, userID, agentID); err != nil {
		return nil, err
	}
	toolType, err := agentToolType(ctx, userID, agentID)
	if err != nil {
		return nil, err
	}
	sessionID := newChatSessionID()
	_, err = dao.AgentChatSessions.Ctx(ctx).Data(do.AgentChatSessions{
		Id:                 sessionID,
		AgentId:            agentID,
		Title:              defaultSessionTitle,
		Status:             SessionStatusIdle,
		ToolType:           toolType,
		RuntimeState:       RuntimeStateIdle,
		MessageCount:       int64(0),
		LastMessagePreview: "",
	}).Insert()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeChatStoreUnavailable)
	}
	return s.GetSession(ctx, userID, agentID, sessionID)
}

// GetSession returns one current-user-visible Chat session.
func (s *serviceImpl) GetSession(ctx context.Context, userID string, agentID string, sessionID string) (*SessionInfo, error) {
	userID, agentID, sessionID = normalizeOwnerAgentSession(userID, agentID, sessionID)
	if err := s.accessSvc.EnsureChatSessionVisible(ctx, userID, agentID, sessionID); err != nil {
		return nil, err
	}
	return getSessionByAgent(ctx, agentID, sessionID)
}

// UpdateSessionTitle updates one visible Chat session title.
func (s *serviceImpl) UpdateSessionTitle(ctx context.Context, userID string, agentID string, sessionID string, title string) (*SessionInfo, error) {
	userID, agentID, sessionID = normalizeOwnerAgentSession(userID, agentID, sessionID)
	if err := s.accessSvc.EnsureChatSessionVisible(ctx, userID, agentID, sessionID); err != nil {
		return nil, err
	}
	title = normalizeTitle(title)
	if title == "" {
		return nil, bizerr.NewCode(CodeChatInvalidInput)
	}
	result, err := dao.AgentChatSessions.Ctx(ctx).
		Where(do.AgentChatSessions{
			Id:      sessionID,
			AgentId: agentID,
		}).
		Data(do.AgentChatSessions{Title: title}).
		Update()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeChatStoreUnavailable)
	}
	if affected, affectedErr := result.RowsAffected(); affectedErr == nil && affected == 0 {
		return nil, bizerr.NewCode(CodeChatNotFound)
	}
	return getSessionByAgent(ctx, agentID, sessionID)
}

// DeleteSession deletes one visible idle Chat session.
func (s *serviceImpl) DeleteSession(ctx context.Context, userID string, agentID string, sessionID string) error {
	userID, agentID, sessionID = normalizeOwnerAgentSession(userID, agentID, sessionID)
	if err := s.accessSvc.EnsureChatSessionVisible(ctx, userID, agentID, sessionID); err != nil {
		return err
	}
	session, err := getSessionByAgent(ctx, agentID, sessionID)
	if err != nil {
		return err
	}
	if session.RuntimeState == RuntimeStateRecovering || session.RuntimeState == RuntimeStateRunning || session.RuntimeState == RuntimeStateWaiting {
		return bizerr.NewCode(CodeChatStateConflict)
	}
	result, err := dao.AgentChatSessions.Ctx(ctx).
		Where(do.AgentChatSessions{
			Id:      sessionID,
			AgentId: agentID,
		}).
		Delete()
	if err != nil {
		return bizerr.WrapCode(err, CodeChatStoreUnavailable)
	}
	if affected, affectedErr := result.RowsAffected(); affectedErr == nil && affected == 0 {
		return bizerr.NewCode(CodeChatNotFound)
	}
	return nil
}

// Messages returns one visible Chat session and its persisted messages.
func (s *serviceImpl) Messages(ctx context.Context, userID string, agentID string, sessionID string) (*MessagesOutput, error) {
	userID, agentID, sessionID = normalizeOwnerAgentSession(userID, agentID, sessionID)
	if err := s.accessSvc.EnsureChatSessionVisible(ctx, userID, agentID, sessionID); err != nil {
		return nil, err
	}
	session, err := getSessionByAgent(ctx, agentID, sessionID)
	if err != nil {
		return nil, err
	}
	var rows []*entity.AgentChatMessages
	cols := dao.AgentChatMessages.Columns()
	err = dao.AgentChatMessages.Ctx(ctx).
		Where(do.AgentChatMessages{SessionId: sessionID}).
		OrderAsc(cols.Sequence).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeChatStoreUnavailable)
	}
	messages := make([]MessageInfo, 0, len(rows))
	for _, row := range rows {
		messages = append(messages, messageFromEntity(row))
	}
	return &MessagesOutput{Session: *session, Messages: messages}, nil
}

// Recover validates visibility but currently waits for runtime migration.
func (s *serviceImpl) Recover(ctx context.Context, userID string, agentID string, sessionID string) (*RecoverOutput, error) {
	if _, err := s.GetSession(ctx, userID, agentID, sessionID); err != nil {
		return nil, err
	}
	return nil, bizerr.NewCode(CodeChatRuntimeUnavailable)
}

func getSessionByAgent(ctx context.Context, agentID string, sessionID string) (*SessionInfo, error) {
	var row *entity.AgentChatSessions
	err := dao.AgentChatSessions.Ctx(ctx).
		Where(do.AgentChatSessions{
			Id:      strings.TrimSpace(sessionID),
			AgentId: strings.TrimSpace(agentID),
		}).
		Scan(&row)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeChatStoreUnavailable)
	}
	if row == nil {
		return nil, bizerr.NewCode(CodeChatNotFound)
	}
	item := sessionFromEntity(row)
	return &item, nil
}

func agentToolType(ctx context.Context, userID string, agentID string) (string, error) {
	var row *entity.CodingAgents
	err := dao.CodingAgents.Ctx(ctx).
		Fields(dao.CodingAgents.Columns().AgentType).
		Where(do.CodingAgents{
			Id:     strings.TrimSpace(agentID),
			UserId: strings.TrimSpace(userID),
		}).
		Scan(&row)
	if err != nil {
		return "", bizerr.WrapCode(err, CodeChatStoreUnavailable)
	}
	if row == nil {
		return "", bizerr.NewCode(CodeChatNotFound)
	}
	return strings.TrimSpace(row.AgentType), nil
}

func sessionFromEntity(row *entity.AgentChatSessions) SessionInfo {
	if row == nil {
		return SessionInfo{}
	}
	return SessionInfo{
		ID:                 row.Id,
		AgentID:            row.AgentId,
		Title:              row.Title,
		Status:             normalizeSessionStatus(row.Status),
		ToolType:           row.ToolType,
		ToolSessionID:      row.ToolSessionId,
		RuntimeState:       normalizeRuntimeState(row.RuntimeState),
		LastError:          row.LastError,
		MessageCount:       row.MessageCount,
		LastMessagePreview: row.LastMessagePreview,
		CreatedAt:          unixMilliFromTimePtr(row.CreatedAt),
		UpdatedAt:          unixMilliFromTimePtr(row.UpdatedAt),
		LastActiveAt:       unixMilliFromTimePtr(row.LastActiveAt),
	}
}

func messageFromEntity(row *entity.AgentChatMessages) MessageInfo {
	if row == nil {
		return MessageInfo{}
	}
	return MessageInfo{
		ID:        row.Id,
		SessionID: row.SessionId,
		Sequence:  row.Sequence,
		Role:      normalizeMessageRole(row.Role),
		Content:   row.Content,
		Status:    normalizeMessageStatus(row.Status),
		Metadata:  row.Metadata,
		CreatedAt: unixMilliFromTimePtr(row.CreatedAt),
		UpdatedAt: unixMilliFromTimePtr(row.UpdatedAt),
	}
}

func normalizeOwnerAndAgent(userID string, agentID string) (string, string) {
	return strings.TrimSpace(userID), strings.TrimSpace(agentID)
}

func normalizeOwnerAgentSession(userID string, agentID string, sessionID string) (string, string, string) {
	return strings.TrimSpace(userID), strings.TrimSpace(agentID), strings.TrimSpace(sessionID)
}
