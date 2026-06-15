// This file implements terminal-session metadata persistence. Agent ownership
// is checked before any terminal row is read or written, and backend session
// names are generated from hashed identifiers instead of raw browser input.

package terminal

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/google/uuid"

	"lina-core/pkg/bizerr"

	"john-ai-agentbox/backend/internal/dao"
	"john-ai-agentbox/backend/internal/model/do"
	"john-ai-agentbox/backend/internal/model/entity"
)

// ListSessions returns persisted terminal sessions for one visible Agent.
func (s *serviceImpl) ListSessions(ctx context.Context, userID string, agentID string, filter SessionFilter) ([]SessionInfo, error) {
	userID, agentID = normalizeOwnerAndAgent(userID, agentID)
	if err := s.accessSvc.EnsureAgentVisible(ctx, userID, agentID); err != nil {
		return nil, err
	}
	status := normalizeStatus(filter.Status)
	if strings.TrimSpace(filter.Status) != "" && status == "" {
		return nil, bizerr.NewCode(CodeTerminalInvalidInput)
	}
	model := dao.AgentTerminalSessions.Ctx(ctx).
		Where(do.AgentTerminalSessions{
			UserId:  userID,
			AgentId: agentID,
		})
	if status != "" {
		model = model.Where(do.AgentTerminalSessions{Status: status})
	}
	cols := dao.AgentTerminalSessions.Columns()
	var rows []*entity.AgentTerminalSessions
	err := model.OrderDesc(cols.UpdatedAt).Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeTerminalStoreUnavailable)
	}
	items := make([]SessionInfo, 0, len(rows))
	for _, row := range rows {
		items = append(items, sessionFromEntity(row))
	}
	return items, nil
}

// EnsureSession creates or rebuilds terminal metadata for one visible Agent.
func (s *serviceImpl) EnsureSession(ctx context.Context, userID string, agentID string, input SessionInput) (*SessionInfo, error) {
	userID, agentID = normalizeOwnerAndAgent(userID, agentID)
	terminalID := strings.TrimSpace(input.TerminalID)
	if userID == "" || agentID == "" || terminalID == "" {
		return nil, bizerr.NewCode(CodeTerminalInvalidInput)
	}
	if err := s.accessSvc.EnsureAgentVisible(ctx, userID, agentID); err != nil {
		return nil, err
	}
	current, err := getSession(ctx, userID, agentID, terminalID)
	if err == nil {
		if current.Status == SessionStatusClosed && !input.Rebuild {
			return nil, bizerr.NewCode(CodeTerminalStateConflict)
		}
		return resetSession(ctx, current.ID, normalizeSessionInput(userID, agentID, input))
	}
	if !bizerr.Is(err, CodeTerminalNotFound) {
		return nil, err
	}
	normalized := normalizeSessionInput(userID, agentID, input)
	sessionID := newSessionID()
	_, err = dao.AgentTerminalSessions.Ctx(ctx).Data(do.AgentTerminalSessions{
		Id:                 sessionID,
		UserId:             userID,
		AgentId:            agentID,
		TerminalId:         terminalID,
		BackendType:        normalized.BackendType,
		BackendSessionName: normalized.BackendSessionName,
		WorkingDir:         normalized.WorkingDir,
		Shell:              normalized.Shell,
		Status:             SessionStatusActive,
		LastError:          "",
	}).Insert()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeTerminalStoreUnavailable)
	}
	return getSession(ctx, userID, agentID, terminalID)
}

// GetSession returns one visible terminal session by browser terminal ID.
func (s *serviceImpl) GetSession(ctx context.Context, userID string, agentID string, terminalID string) (*SessionInfo, error) {
	userID, agentID = normalizeOwnerAndAgent(userID, agentID)
	terminalID = strings.TrimSpace(terminalID)
	if err := s.accessSvc.EnsureTerminalSessionVisible(ctx, userID, agentID, terminalID); err != nil {
		return nil, err
	}
	return getSession(ctx, userID, agentID, terminalID)
}

// CloseSession marks one visible terminal session closed.
func (s *serviceImpl) CloseSession(ctx context.Context, userID string, agentID string, terminalID string) error {
	userID, agentID = normalizeOwnerAndAgent(userID, agentID)
	terminalID = strings.TrimSpace(terminalID)
	if err := s.accessSvc.EnsureTerminalSessionVisible(ctx, userID, agentID, terminalID); err != nil {
		return err
	}
	cols := dao.AgentTerminalSessions.Columns()
	closedAt := time.Now()
	result, err := dao.AgentTerminalSessions.Ctx(ctx).
		Where(do.AgentTerminalSessions{
			UserId:     userID,
			AgentId:    agentID,
			TerminalId: terminalID,
		}).
		WhereNot(cols.Status, SessionStatusClosed).
		Data(do.AgentTerminalSessions{
			Status:    SessionStatusClosed,
			LastError: "",
			ClosedAt:  &closedAt,
		}).
		Update()
	if err != nil {
		return bizerr.WrapCode(err, CodeTerminalStoreUnavailable)
	}
	if affected, affectedErr := result.RowsAffected(); affectedErr == nil && affected == 0 {
		if _, getErr := getSession(ctx, userID, agentID, terminalID); getErr != nil {
			return getErr
		}
	}
	return nil
}

type normalizedInput struct {
	BackendType        string
	BackendSessionName string
	WorkingDir         string
	Shell              string
}

func getSession(ctx context.Context, userID string, agentID string, terminalID string) (*SessionInfo, error) {
	var row *entity.AgentTerminalSessions
	err := dao.AgentTerminalSessions.Ctx(ctx).
		Where(do.AgentTerminalSessions{
			UserId:     strings.TrimSpace(userID),
			AgentId:    strings.TrimSpace(agentID),
			TerminalId: strings.TrimSpace(terminalID),
		}).
		Scan(&row)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeTerminalStoreUnavailable)
	}
	if row == nil {
		return nil, bizerr.NewCode(CodeTerminalNotFound)
	}
	item := sessionFromEntity(row)
	return &item, nil
}

func resetSession(ctx context.Context, sessionID string, input normalizedInput) (*SessionInfo, error) {
	sessionID = strings.TrimSpace(sessionID)
	cols := dao.AgentTerminalSessions.Columns()
	result, err := dao.AgentTerminalSessions.Ctx(ctx).
		Where(do.AgentTerminalSessions{Id: strings.TrimSpace(sessionID)}).
		Data(do.AgentTerminalSessions{
			BackendType:        input.BackendType,
			BackendSessionName: input.BackendSessionName,
			WorkingDir:         input.WorkingDir,
			Shell:              input.Shell,
			Status:             SessionStatusActive,
			LastError:          "",
		}).
		Update()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeTerminalStoreUnavailable)
	}
	if affected, affectedErr := result.RowsAffected(); affectedErr == nil && affected == 0 {
		return nil, bizerr.NewCode(CodeTerminalNotFound)
	}
	_, err = dao.AgentTerminalSessions.Ctx(ctx).
		Where(do.AgentTerminalSessions{Id: strings.TrimSpace(sessionID)}).
		Data(cols.ClosedAt, gdb.Raw("NULL")).
		Update()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeTerminalStoreUnavailable)
	}
	return getSessionByID(ctx, sessionID)
}

func getSessionByID(ctx context.Context, sessionID string) (*SessionInfo, error) {
	var row *entity.AgentTerminalSessions
	err := dao.AgentTerminalSessions.Ctx(ctx).
		Where(do.AgentTerminalSessions{Id: strings.TrimSpace(sessionID)}).
		Scan(&row)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeTerminalStoreUnavailable)
	}
	if row == nil {
		return nil, bizerr.NewCode(CodeTerminalNotFound)
	}
	item := sessionFromEntity(row)
	return &item, nil
}

func normalizeOwnerAndAgent(userID string, agentID string) (string, string) {
	return strings.TrimSpace(userID), strings.TrimSpace(agentID)
}

func normalizeSessionInput(userID string, agentID string, input SessionInput) normalizedInput {
	workingDir := strings.TrimSpace(input.WorkingDir)
	if workingDir == "" {
		workingDir = DefaultWorkingDir
	}
	shell := strings.TrimSpace(input.Shell)
	if shell == "" {
		shell = DefaultShell
	}
	return normalizedInput{
		BackendType:        BackendTypeTmux,
		BackendSessionName: backendSessionName(userID, agentID, input.TerminalID),
		WorkingDir:         workingDir,
		Shell:              shell,
	}
}

func normalizeStatus(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "":
		return ""
	case SessionStatusActive:
		return SessionStatusActive
	case SessionStatusClosed:
		return SessionStatusClosed
	case SessionStatusError:
		return SessionStatusError
	default:
		return ""
	}
}

func sessionFromEntity(row *entity.AgentTerminalSessions) SessionInfo {
	if row == nil {
		return SessionInfo{}
	}
	return SessionInfo{
		ID:                 row.Id,
		UserID:             row.UserId,
		AgentID:            row.AgentId,
		TerminalID:         row.TerminalId,
		BackendType:        row.BackendType,
		BackendSessionName: row.BackendSessionName,
		WorkingDir:         row.WorkingDir,
		Shell:              row.Shell,
		Status:             normalizeStatus(row.Status),
		LastError:          row.LastError,
		ClosedAt:           unixMilliPtrFromTimePtr(row.ClosedAt),
		CreatedAt:          unixMilliFromTimePtr(row.CreatedAt),
		UpdatedAt:          unixMilliFromTimePtr(row.UpdatedAt),
	}
}

func backendSessionName(userID string, agentID string, terminalID string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(userID) + "\x00" + strings.TrimSpace(agentID) + "\x00" + strings.TrimSpace(terminalID)))
	return "abx-" + hex.EncodeToString(sum[:])[:24]
}

func newSessionID() string {
	return "term-" + strings.ReplaceAll(uuid.NewString(), "-", "")
}
