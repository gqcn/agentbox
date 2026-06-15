// This file centralizes AgentBox session cookie access at the HTTP controller
// boundary. The service layer receives only opaque token values so it stays
// independent from GoFrame request objects.

package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

const agentBoxSessionCookieName = "agent_box_session"

// sessionTokenFromContext reads the independent AgentBox session cookie from
// the current HTTP request. Missing request context is treated as anonymous.
func sessionTokenFromContext(ctx context.Context) string {
	request := g.RequestFromCtx(ctx)
	if request == nil || request.Cookie == nil {
		return ""
	}
	return request.Cookie.Get(agentBoxSessionCookieName).String()
}

// setSessionCookie writes the independent AgentBox session cookie when the
// authentication service returns a new opaque token.
func setSessionCookie(ctx context.Context, token string, expiresAtMillis int64) {
	if token == "" || expiresAtMillis <= 0 {
		return
	}
	request := g.RequestFromCtx(ctx)
	if request == nil || request.Cookie == nil {
		return
	}
	maxAge := time.Until(time.UnixMilli(expiresAtMillis))
	if maxAge <= 0 {
		maxAge = time.Second
	}
	request.Cookie.SetCookie(
		agentBoxSessionCookieName,
		token,
		"",
		"/",
		maxAge,
		ghttp.CookieOptions{
			SameSite: http.SameSiteLaxMode,
			HttpOnly: true,
		},
	)
}

// clearSessionCookie expires the independent AgentBox browser-session cookie
// without touching LinaPro management-workbench authentication cookies.
func clearSessionCookie(ctx context.Context) {
	request := g.RequestFromCtx(ctx)
	if request == nil || request.Cookie == nil {
		return
	}
	request.Cookie.SetCookie(
		agentBoxSessionCookieName,
		"",
		"",
		"/",
		-24*time.Hour,
		ghttp.CookieOptions{
			SameSite: http.SameSiteLaxMode,
			HttpOnly: true,
		},
	)
}
