// This file reads the current independent AgentBox browser session from the
// plugin-owned session cookie.

package auth

import (
	"context"

	v1 "john-ai-agentbox/backend/api/auth/v1"
)

// Session returns the current AgentBox browser session.
func (c *ControllerV1) Session(ctx context.Context, _ *v1.SessionReq) (res *v1.SessionRes, err error) {
	out, err := c.authSvc.CurrentSession(ctx, sessionTokenFromContext(ctx))
	if err != nil {
		return nil, err
	}
	return toAuthSessionResponse(out), nil
}
