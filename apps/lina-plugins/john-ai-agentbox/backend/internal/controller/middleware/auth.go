// This file implements AgentBox protected-route authentication middleware. It
// validates only the plugin-owned agent_box_session cookie and writes the
// resolved AgentBox user into request context for downstream ownership checks.

package middleware

import (
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/bizerr"

	authsvc "john-ai-agentbox/backend/internal/service/auth"
	"john-ai-agentbox/backend/internal/service/authctx"
)

const agentBoxSessionCookieName = "agent_box_session"

// Auth returns a middleware that rejects requests without a valid AgentBox
// session and propagates the authenticated plugin user through context.
func Auth(authSvc authsvc.Service) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		if authSvc == nil {
			r.SetError(bizerr.NewCode(authsvc.CodeAuthStoreUnavailable))
			return
		}
		token := r.Cookie.Get(agentBoxSessionCookieName, "").String()
		session, err := authSvc.CurrentSession(r.Context(), token)
		if err != nil {
			r.SetError(err)
			return
		}
		if session == nil || session.User == nil {
			r.SetError(bizerr.NewCode(authsvc.CodeAuthRequired))
			return
		}
		r.SetCtx(authctx.WithUser(r.Context(), *session.User))
		r.Middleware.Next()
	}
}
