// This file verifies AgentBox authentication service behavior without a
// database. The fake store preserves the production boundary: callers provide
// only opaque tokens while storage records only SHA-256 token hashes.

package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"lina-core/pkg/bizerr"

	"golang.org/x/crypto/bcrypt"
)

func TestServiceLoginSessionAndLogout(t *testing.T) {
	ctx := context.Background()
	store := newAuthServiceTestStore(t)
	service, err := New(store, Config{SessionTTL: time.Hour})
	if err != nil {
		t.Fatal(err)
	}

	out, err := service.Login(ctx, LoginInput{
		Username:  " admin ",
		Password:  "admin123",
		UserAgent: "Go test",
		ClientIP:  "127.0.0.1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if out == nil || out.User == nil || out.User.ID != "usr-admin" {
		t.Fatalf("login output = %#v", out)
	}
	if out.Token == "" {
		t.Fatal("expected opaque token")
	}
	if store.createdTokenHash == "" || store.createdTokenHash == out.Token {
		t.Fatalf("stored token hash = %q token = %q", store.createdTokenHash, out.Token)
	}
	if store.createdTokenHash != sessionTokenHash(out.Token) {
		t.Fatalf("stored token hash = %q, want %q", store.createdTokenHash, sessionTokenHash(out.Token))
	}
	if store.createdUserAgent != "Go test" || store.createdClientIP != "127.0.0.1" {
		t.Fatalf("stored session request metadata ua=%q ip=%q", store.createdUserAgent, store.createdClientIP)
	}
	if store.touchedUserID != "usr-admin" {
		t.Fatalf("touched user id = %q", store.touchedUserID)
	}

	session, err := service.CurrentSession(ctx, out.Token)
	if err != nil {
		t.Fatal(err)
	}
	if session == nil || session.User.Username != "admin" {
		t.Fatalf("session output = %#v", session)
	}

	logout, err := service.Logout(ctx, out.Token)
	if err != nil {
		t.Fatal(err)
	}
	if logout == nil || !logout.LoggedOut || !store.revoked[sessionTokenHash(out.Token)] {
		t.Fatalf("logout = %#v revoked=%v", logout, store.revoked)
	}
	_, err = service.CurrentSession(ctx, out.Token)
	if !bizerr.Is(err, CodeAuthRequired) {
		t.Fatalf("expected auth required after logout, got %v", err)
	}
}

func TestServiceRejectsInvalidCredentials(t *testing.T) {
	ctx := context.Background()
	store := newAuthServiceTestStore(t)
	service, err := New(store, Config{SessionTTL: time.Hour})
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.Login(ctx, LoginInput{Username: "admin", Password: "wrong-password"})
	if !bizerr.Is(err, CodeAuthInvalidCredentials) {
		t.Fatalf("expected invalid credentials for wrong password, got %v", err)
	}
	_, err = service.Login(ctx, LoginInput{Username: "missing", Password: "admin123"})
	if !bizerr.Is(err, CodeAuthInvalidCredentials) {
		t.Fatalf("expected invalid credentials for missing user, got %v", err)
	}
	if len(store.sessions) != 0 {
		t.Fatalf("sessions created after failed login: %#v", store.sessions)
	}
}

func TestServiceRejectsDisabledUser(t *testing.T) {
	ctx := context.Background()
	store := newAuthServiceTestStore(t)
	store.user.Status = UserStatusDisabled
	service, err := New(store, Config{SessionTTL: time.Hour})
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.Login(ctx, LoginInput{Username: "admin", Password: "admin123"})
	if !bizerr.Is(err, CodeAuthUserDisabled) {
		t.Fatalf("expected disabled user error, got %v", err)
	}
}

func TestServiceRejectsExpiredSession(t *testing.T) {
	ctx := context.Background()
	store := newAuthServiceTestStore(t)
	service, err := New(store, Config{SessionTTL: time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	token := "expired-token"
	hash := sessionTokenHash(token)
	store.sessions[hash] = &SessionRecord{
		TokenHash: hash,
		UserID:    "usr-admin",
		ExpiresAt: time.Now().Add(-time.Minute).UnixMilli(),
		CreatedAt: time.Now().Add(-2 * time.Minute).UnixMilli(),
		UpdatedAt: time.Now().Add(-2 * time.Minute).UnixMilli(),
	}

	_, err = service.CurrentSession(ctx, token)
	if !bizerr.Is(err, CodeAuthRequired) {
		t.Fatalf("expected auth required for expired token, got %v", err)
	}
}

func TestServiceLogoutWithoutTokenIsIdempotent(t *testing.T) {
	service, err := New(newAuthServiceTestStore(t), Config{SessionTTL: time.Hour})
	if err != nil {
		t.Fatal(err)
	}

	out, err := service.Logout(context.Background(), " ")
	if err != nil {
		t.Fatal(err)
	}
	if out == nil || !out.LoggedOut {
		t.Fatalf("logout output = %#v", out)
	}
}

type authServiceTestStore struct {
	user              UserRecord
	sessions          map[string]*SessionRecord
	revoked           map[string]bool
	createdTokenHash  string
	createdUserAgent  string
	createdClientIP   string
	touchedUserID     string
	touchErr          error
	createSessionErr  error
	getUserByUsername map[string]*UserRecord
}

func newAuthServiceTestStore(t *testing.T) *authServiceTestStore {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UnixMilli()
	user := UserRecord{
		UserOutput: UserOutput{
			ID:          "usr-admin",
			Username:    "admin",
			DisplayName: "admin",
			LastLoginAt: 0,
		},
		PasswordHash: string(hash),
		Role:         "admin",
		Status:       UserStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	return &authServiceTestStore{
		user:     user,
		sessions: make(map[string]*SessionRecord),
		revoked:  make(map[string]bool),
	}
}

func (s *authServiceTestStore) GetUserByUsername(_ context.Context, username string) (*UserRecord, error) {
	if s.getUserByUsername != nil {
		if user, ok := s.getUserByUsername[strings.TrimSpace(username)]; ok {
			copy := *user
			return &copy, nil
		}
		return nil, errAuthRecordNotFound
	}
	if strings.TrimSpace(username) != "admin" {
		return nil, errAuthRecordNotFound
	}
	copy := s.user
	return &copy, nil
}

func (s *authServiceTestStore) TouchUserLogin(_ context.Context, userID string) error {
	if s.touchErr != nil {
		return s.touchErr
	}
	s.touchedUserID = userID
	s.user.LastLoginAt = time.Now().UnixMilli()
	return nil
}

func (s *authServiceTestStore) CreateUserSession(_ context.Context, tokenHash string, userID string, userAgent string, clientIP string, expiresAt time.Time) (*SessionRecord, error) {
	if s.createSessionErr != nil {
		return nil, s.createSessionErr
	}
	session := &SessionRecord{
		TokenHash: tokenHash,
		UserID:    userID,
		ExpiresAt: expiresAt.UnixMilli(),
		CreatedAt: time.Now().UnixMilli(),
		UpdatedAt: time.Now().UnixMilli(),
	}
	s.createdTokenHash = tokenHash
	s.createdUserAgent = userAgent
	s.createdClientIP = clientIP
	s.sessions[tokenHash] = session
	return session, nil
}

func (s *authServiceTestStore) GetValidUserSession(_ context.Context, tokenHash string, now time.Time) (*UserRecord, *SessionRecord, error) {
	session, ok := s.sessions[tokenHash]
	if !ok || s.revoked[tokenHash] || session.ExpiresAt <= now.UnixMilli() {
		return nil, nil, errAuthRecordNotFound
	}
	user := s.user
	if user.Status != UserStatusActive {
		return nil, nil, errAuthRecordNotFound
	}
	return &user, session, nil
}

func (s *authServiceTestStore) RevokeUserSession(_ context.Context, tokenHash string) error {
	if tokenHash == "revoke-error" {
		return errors.New("revoke failed")
	}
	s.revoked[tokenHash] = true
	return nil
}
