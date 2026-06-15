// This file implements the DAO-backed AgentBox authentication store. It uses
// generated DAO/DO/Entity objects over plugin-owned john_ai_agentbox_* tables so
// authentication storage remains isolated from LinaPro management users and sessions.

package auth

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"john-ai-agentbox/backend/internal/dao"
	"john-ai-agentbox/backend/internal/model/do"
	"john-ai-agentbox/backend/internal/model/entity"
)

// daoStore persists AgentBox auth state in plugin-owned tables.
type daoStore struct{}

var _ Store = (*daoStore)(nil)

// NewDAOStore creates the production AgentBox auth store backed by generated DAO.
func NewDAOStore() Store {
	return &daoStore{}
}

// GetUserByUsername returns one non-deleted AgentBox user by login username.
func (s *daoStore) GetUserByUsername(ctx context.Context, username string) (*UserRecord, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, errAuthRecordNotFound
	}
	cols := dao.Users.Columns()
	var row *entity.Users
	err := dao.Users.Ctx(ctx).
		Wheref("LOWER(%s)=LOWER(?)", cols.Username, username).
		Scan(&row)
	if err != nil {
		return nil, gerror.Wrap(err, "query agentbox user by username")
	}
	if row == nil {
		return nil, errAuthRecordNotFound
	}
	return userRecordFromEntity(row), nil
}

// TouchUserLogin records the latest successful AgentBox login timestamp.
func (s *daoStore) TouchUserLogin(ctx context.Context, userID string) error {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return errAuthRecordNotFound
	}
	now := time.Now()
	result, err := dao.Users.Ctx(ctx).
		Where(do.Users{Id: userID}).
		Data(do.Users{LastLoginAt: &now}).
		Update()
	if err != nil {
		return gerror.Wrap(err, "touch agentbox user login")
	}
	if affected, affectedErr := result.RowsAffected(); affectedErr == nil && affected == 0 {
		return errAuthRecordNotFound
	}
	return nil
}

// CreateUserSession persists one hashed AgentBox browser session.
func (s *daoStore) CreateUserSession(
	ctx context.Context,
	tokenHash string,
	userID string,
	userAgent string,
	clientIP string,
	expiresAt time.Time,
) (*SessionRecord, error) {
	tokenHash = strings.TrimSpace(tokenHash)
	userID = strings.TrimSpace(userID)
	if tokenHash == "" || userID == "" || expiresAt.IsZero() {
		return nil, errAuthRecordNotFound
	}
	_, err := dao.UserSessions.Ctx(ctx).Data(do.UserSessions{
		TokenHash: tokenHash,
		UserId:    userID,
		UserAgent: strings.TrimSpace(userAgent),
		IpAddress: strings.TrimSpace(clientIP),
		ExpiresAt: &expiresAt,
	}).Insert()
	if err != nil {
		return nil, gerror.Wrap(err, "create agentbox user session")
	}
	return &SessionRecord{
		TokenHash: tokenHash,
		UserID:    userID,
		ExpiresAt: expiresAt.UnixMilli(),
		CreatedAt: time.Now().UnixMilli(),
		UpdatedAt: time.Now().UnixMilli(),
	}, nil
}

// GetValidUserSession resolves one non-revoked, non-expired AgentBox session.
func (s *daoStore) GetValidUserSession(ctx context.Context, tokenHash string, now time.Time) (*UserRecord, *SessionRecord, error) {
	tokenHash = strings.TrimSpace(tokenHash)
	if tokenHash == "" {
		return nil, nil, errAuthRecordNotFound
	}
	if now.IsZero() {
		now = time.Now()
	}
	session, err := s.getActiveSession(ctx, tokenHash, now)
	if err != nil {
		return nil, nil, err
	}
	user, err := s.getActiveUserByID(ctx, session.UserId)
	if err != nil {
		return nil, nil, err
	}
	return user, sessionRecordFromEntity(session), nil
}

// RevokeUserSession marks one AgentBox session hash unusable.
func (s *daoStore) RevokeUserSession(ctx context.Context, tokenHash string) error {
	tokenHash = strings.TrimSpace(tokenHash)
	if tokenHash == "" {
		return nil
	}
	now := time.Now()
	_, err := dao.UserSessions.Ctx(ctx).
		Where(do.UserSessions{TokenHash: tokenHash}).
		WhereNull(dao.UserSessions.Columns().RevokedAt).
		Data(do.UserSessions{RevokedAt: &now}).
		Update()
	return gerror.Wrap(err, "revoke agentbox user session")
}

// getActiveSession reads one unrevoked and unexpired session row.
func (s *daoStore) getActiveSession(ctx context.Context, tokenHash string, now time.Time) (*entity.UserSessions, error) {
	cols := dao.UserSessions.Columns()
	var row *entity.UserSessions
	err := dao.UserSessions.Ctx(ctx).
		Where(do.UserSessions{TokenHash: tokenHash}).
		WhereNull(cols.RevokedAt).
		WhereGT(cols.ExpiresAt, now).
		Scan(&row)
	if err != nil {
		return nil, gerror.Wrap(err, "query agentbox user session")
	}
	if row == nil {
		return nil, errAuthRecordNotFound
	}
	return row, nil
}

// getActiveUserByID reads one active user for session authentication.
func (s *daoStore) getActiveUserByID(ctx context.Context, userID string) (*UserRecord, error) {
	var row *entity.Users
	err := dao.Users.Ctx(ctx).
		Where(do.Users{Id: strings.TrimSpace(userID), Status: UserStatusActive}).
		Scan(&row)
	if err != nil {
		return nil, gerror.Wrap(err, "query active agentbox user")
	}
	if row == nil {
		return nil, errAuthRecordNotFound
	}
	return userRecordFromEntity(row), nil
}

// userRecordFromEntity maps a generated user entity into the auth store record.
func userRecordFromEntity(row *entity.Users) *UserRecord {
	if row == nil {
		return nil
	}
	return &UserRecord{
		UserOutput: UserOutput{
			ID:          row.Id,
			Username:    row.Username,
			DisplayName: row.Username,
			LastLoginAt: unixMilliFromTimePtr(row.LastLoginAt),
		},
		PasswordHash: row.PasswordHash,
		Role:         row.Role,
		Status:       row.Status,
		CreatedAt:    unixMilliFromTimePtr(row.CreatedAt),
		UpdatedAt:    unixMilliFromTimePtr(row.UpdatedAt),
	}
}

// sessionRecordFromEntity maps a generated session entity into the auth record.
func sessionRecordFromEntity(row *entity.UserSessions) *SessionRecord {
	if row == nil {
		return nil
	}
	return &SessionRecord{
		TokenHash: row.TokenHash,
		UserID:    row.UserId,
		ExpiresAt: unixMilliFromTimePtr(row.ExpiresAt),
		CreatedAt: unixMilliFromTimePtr(row.CreatedAt),
		UpdatedAt: unixMilliFromTimePtr(row.UpdatedAt),
	}
}

// unixMilliFromTimePtr converts nullable database timestamps for API projection.
func unixMilliFromTimePtr(value *time.Time) int64 {
	if value == nil || value.IsZero() {
		return 0
	}
	return value.UnixMilli()
}
