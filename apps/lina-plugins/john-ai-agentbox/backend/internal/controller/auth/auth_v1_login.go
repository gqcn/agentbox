// This file handles AgentBox login requests and keeps response projection
// separate from the plugin-owned authentication service.

package auth

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	v1 "john-ai-agentbox/backend/api/auth/v1"
	authsvc "john-ai-agentbox/backend/internal/service/auth"
)

// Login creates an AgentBox browser session.
func (c *ControllerV1) Login(ctx context.Context, req *v1.LoginReq) (res *v1.LoginRes, err error) {
	var (
		userAgent string
		clientIP  string
	)
	if request := g.RequestFromCtx(ctx); request != nil {
		userAgent = request.UserAgent()
		clientIP = request.GetClientIp()
	}
	out, err := c.authSvc.Login(ctx, authsvc.LoginInput{
		Username:  req.Username,
		Password:  req.Password,
		UserAgent: userAgent,
		ClientIP:  clientIP,
	})
	if err != nil {
		return nil, err
	}
	setSessionCookie(ctx, out.Token, out.ExpiresAt)
	return toAuthSessionResponse(out), nil
}
