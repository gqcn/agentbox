// This file defines the public authentication API contract surface for the
// AgentBox plugin. DTOs live under versioned subpackages while this package
// exposes the GoFrame controller interface used by route registration.

package auth

import (
	"context"

	v1 "john-ai-agentbox/backend/api/auth/v1"
)

// IAuthV1 defines AgentBox browser-session HTTP handlers.
type IAuthV1 interface {
	// Login creates one AgentBox session from username and password credentials.
	Login(ctx context.Context, req *v1.LoginReq) (res *v1.LoginRes, err error)
	// Session returns the authenticated AgentBox session bound to the request.
	Session(ctx context.Context, req *v1.SessionReq) (res *v1.SessionRes, err error)
	// Logout revokes the current AgentBox session.
	Logout(ctx context.Context, req *v1.LogoutReq) (res *v1.LogoutRes, err error)
}
