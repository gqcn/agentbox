// This file handles idempotent AgentBox logout requests for the independent
// plugin session cookie.

package auth

import (
	"context"

	v1 "john-ai-agentbox/backend/api/auth/v1"
)

// Logout revokes the current AgentBox browser session.
func (c *ControllerV1) Logout(ctx context.Context, _ *v1.LogoutReq) (res *v1.LogoutRes, err error) {
	out, err := c.authSvc.Logout(ctx, sessionTokenFromContext(ctx))
	if err != nil {
		return nil, err
	}
	clearSessionCookie(ctx)
	return &v1.AuthLogoutResponse{LoggedOut: out.LoggedOut}, nil
}
