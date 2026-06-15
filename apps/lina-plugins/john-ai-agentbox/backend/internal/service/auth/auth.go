// Package auth owns AgentBox's independent browser-session service contract.
// It verifies plugin-owned users, stores only session token hashes, and keeps
// AgentBox browser sessions independent from LinaPro management authentication.
package auth

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
)

const (
	// defaultSessionTTL is the initial AgentBox browser session duration.
	defaultSessionTTL = 24 * time.Hour
	// UserStatusActive allows an AgentBox user to authenticate.
	UserStatusActive = "active"
	// UserStatusDisabled prevents an AgentBox user from creating or using sessions.
	UserStatusDisabled = "disabled"
)

var errAuthRecordNotFound = gerror.New("agentbox auth record not found")

// Config contains pure value settings for AgentBox authentication.
type Config struct {
	// SessionTTL controls how long one opaque browser session remains valid.
	// A zero value uses the default 24-hour TTL.
	SessionTTL time.Duration
}

// LoginInput carries AgentBox username and password credentials.
type LoginInput struct {
	Username  string
	Password  string
	UserAgent string
	ClientIP  string
}

// Service defines AgentBox's independent authentication boundary.
type Service interface {
	// Login verifies AgentBox credentials and creates a new independent
	// browser session. It returns a public user projection plus opaque cookie
	// material and structured bizerr failures for invalid credentials, disabled
	// users, token generation failures, or storage failures.
	Login(ctx context.Context, in LoginInput) (*SessionOutput, error)
	// CurrentSession resolves an opaque AgentBox session token into a user
	// projection. Empty, expired, revoked, or disabled-user sessions return a
	// structured authentication-required bizerr.
	CurrentSession(ctx context.Context, token string) (*SessionOutput, error)
	// Logout revokes an opaque AgentBox session token. Empty tokens are a
	// successful no-op so clients can clear cookies idempotently; storage
	// failures are returned as structured bizerr failures.
	Logout(ctx context.Context, token string) (*LogoutOutput, error)
}

// Store defines the plugin-owned persistence boundary used by auth flows.
type Store interface {
	// GetUserByUsername returns one non-deleted AgentBox user with password hash
	// material for login. Missing users return errAuthRecordNotFound.
	GetUserByUsername(ctx context.Context, username string) (*UserRecord, error)
	// TouchUserLogin records the latest successful login time for the user.
	TouchUserLogin(ctx context.Context, userID string) error
	// CreateUserSession persists a SHA-256 session token hash and returns the
	// server-side session projection. The opaque token itself must never be
	// stored.
	CreateUserSession(ctx context.Context, tokenHash string, userID string, userAgent string, clientIP string, expiresAt time.Time) (*SessionRecord, error)
	// GetValidUserSession resolves a non-revoked, non-expired session hash and
	// its active user. Missing sessions return errAuthRecordNotFound.
	GetValidUserSession(ctx context.Context, tokenHash string, now time.Time) (*UserRecord, *SessionRecord, error)
	// RevokeUserSession marks one session hash unusable. Missing sessions are a
	// successful no-op so logout stays idempotent.
	RevokeUserSession(ctx context.Context, tokenHash string) error
}

// UserRecord includes secret authentication material and must not be serialized.
type UserRecord struct {
	UserOutput
	PasswordHash string
	Role         string
	Status       string
	CreatedAt    int64
	UpdatedAt    int64
}

// SessionRecord is the server-side persisted browser-session projection.
type SessionRecord struct {
	TokenHash string
	UserID    string
	ExpiresAt int64
	CreatedAt int64
	UpdatedAt int64
}

// SessionOutput is the service-level current-session projection.
type SessionOutput struct {
	User      *UserOutput
	Token     string
	ExpiresAt int64
}

// UserOutput is the service-level authenticated AgentBox user projection.
type UserOutput struct {
	ID          string
	Username    string
	DisplayName string
	LastLoginAt int64
}

// LogoutOutput reports whether logout revoked an active AgentBox session.
type LogoutOutput struct {
	LoggedOut bool
}

// serviceImpl is the default AgentBox auth service implementation.
type serviceImpl struct {
	store      Store
	sessionTTL time.Duration
}

var _ Service = (*serviceImpl)(nil)

// New creates the AgentBox authentication service.
func New(store Store, config Config) (Service, error) {
	if store == nil {
		return nil, gerror.New("agentbox auth store is required")
	}
	sessionTTL := config.SessionTTL
	if sessionTTL == 0 {
		sessionTTL = defaultSessionTTL
	}
	if sessionTTL < time.Minute {
		return nil, gerror.New("agentbox auth session ttl must be at least one minute")
	}
	return &serviceImpl{
		store:      store,
		sessionTTL: sessionTTL,
	}, nil
}
