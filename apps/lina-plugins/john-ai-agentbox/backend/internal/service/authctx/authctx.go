// Package authctx stores authenticated AgentBox identity in request contexts.
// Downstream controllers and services use this package to enforce plugin-owned
// user/resource boundaries without parsing cookies in each call path.
package authctx

import (
	"context"

	"lina-core/pkg/bizerr"

	authsvc "john-ai-agentbox/backend/internal/service/auth"
)

type contextKey string

const userContextKey contextKey = "john-ai-agentbox-auth-user"

// WithUser returns a child context carrying the validated AgentBox user.
func WithUser(ctx context.Context, user authsvc.UserOutput) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// User returns the authenticated AgentBox user in ctx.
func User(ctx context.Context) (authsvc.UserOutput, bool) {
	value := ctx.Value(userContextKey)
	user, ok := value.(authsvc.UserOutput)
	if !ok || user.ID == "" {
		return authsvc.UserOutput{}, false
	}
	return user, true
}

// UserID returns the authenticated AgentBox user ID in ctx.
func UserID(ctx context.Context) (string, bool) {
	user, ok := User(ctx)
	if !ok {
		return "", false
	}
	return user.ID, true
}

// RequireUserID returns the authenticated user ID or an auth-required bizerr.
func RequireUserID(ctx context.Context) (string, error) {
	userID, ok := UserID(ctx)
	if !ok {
		return "", bizerr.NewCode(authsvc.CodeAuthRequired)
	}
	return userID, nil
}
