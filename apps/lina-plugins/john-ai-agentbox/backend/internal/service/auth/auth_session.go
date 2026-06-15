// This file implements AgentBox credential verification and session lifecycle.
// Opaque tokens are generated with crypto/rand, while persistent storage only
// receives SHA-256 hashes so database rows cannot be reused as bearer tokens.

package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"strings"
	"time"

	"lina-core/pkg/bizerr"

	"golang.org/x/crypto/bcrypt"
)

const sessionTokenBytes = 32

// Login verifies AgentBox credentials and creates a browser session.
func (s *serviceImpl) Login(ctx context.Context, in LoginInput) (*SessionOutput, error) {
	username := strings.TrimSpace(in.Username)
	if username == "" || in.Password == "" {
		return nil, bizerr.NewCode(CodeAuthInvalidCredentials)
	}
	user, err := s.store.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, errAuthRecordNotFound) {
			return nil, bizerr.NewCode(CodeAuthInvalidCredentials)
		}
		return nil, bizerr.WrapCode(err, CodeAuthStoreUnavailable)
	}
	if user.Status != UserStatusActive {
		return nil, bizerr.NewCode(CodeAuthUserDisabled)
	}
	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)); err != nil {
		return nil, bizerr.NewCode(CodeAuthInvalidCredentials)
	}

	token, err := generateSessionToken()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeAuthStoreUnavailable)
	}
	tokenHash := sessionTokenHash(token)
	expiresAt := time.Now().Add(s.sessionTTL)
	session, err := s.store.CreateUserSession(
		ctx,
		tokenHash,
		user.ID,
		strings.TrimSpace(in.UserAgent),
		strings.TrimSpace(in.ClientIP),
		expiresAt,
	)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeAuthStoreUnavailable)
	}
	if err = s.store.TouchUserLogin(ctx, user.ID); err != nil {
		if cleanupErr := s.store.RevokeUserSession(ctx, tokenHash); cleanupErr != nil {
			return nil, bizerr.WrapCode(cleanupErr, CodeAuthStoreUnavailable)
		}
		return nil, bizerr.WrapCode(err, CodeAuthStoreUnavailable)
	}

	user.LastLoginAt = time.Now().UnixMilli()
	return &SessionOutput{
		User:      publicUser(user),
		Token:     token,
		ExpiresAt: session.ExpiresAt,
	}, nil
}

// CurrentSession resolves an opaque AgentBox session token.
func (s *serviceImpl) CurrentSession(ctx context.Context, token string) (*SessionOutput, error) {
	if strings.TrimSpace(token) == "" {
		return nil, bizerr.NewCode(CodeAuthRequired)
	}
	user, session, err := s.store.GetValidUserSession(ctx, sessionTokenHash(token), time.Now())
	if err != nil {
		if errors.Is(err, errAuthRecordNotFound) {
			return nil, bizerr.NewCode(CodeAuthRequired)
		}
		return nil, bizerr.WrapCode(err, CodeAuthStoreUnavailable)
	}
	return &SessionOutput{
		User:      publicUser(user),
		ExpiresAt: session.ExpiresAt,
	}, nil
}

// Logout revokes an opaque AgentBox session token.
func (s *serviceImpl) Logout(ctx context.Context, token string) (*LogoutOutput, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return &LogoutOutput{LoggedOut: true}, nil
	}
	if err := s.store.RevokeUserSession(ctx, sessionTokenHash(token)); err != nil {
		return nil, bizerr.WrapCode(err, CodeAuthStoreUnavailable)
	}
	return &LogoutOutput{LoggedOut: true}, nil
}

// generateSessionToken creates an opaque URL-safe browser-session token.
func generateSessionToken() (string, error) {
	raw := make([]byte, sessionTokenBytes)
	if _, err := io.ReadFull(rand.Reader, raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// sessionTokenHash returns the stable database key for an opaque token.
func sessionTokenHash(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}

// publicUser strips password hash and storage-only fields from one user record.
func publicUser(user *UserRecord) *UserOutput {
	if user == nil {
		return nil
	}
	displayName := strings.TrimSpace(user.DisplayName)
	if displayName == "" {
		displayName = user.Username
	}
	return &UserOutput{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: displayName,
		LastLoginAt: user.LastLoginAt,
	}
}
