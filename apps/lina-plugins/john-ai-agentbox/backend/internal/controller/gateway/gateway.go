// Package gateway implements AgentBox raw runtime HTTP handlers. Handlers
// resolve the current AgentBox user from request context and delegate
// ownership and runtime readiness checks to the plugin-owned gateway service.
package gateway

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"

	"john-ai-agentbox/backend/internal/service/authctx"
	gatewaysvc "john-ai-agentbox/backend/internal/service/gateway"
)

// Controller handles AgentBox raw runtime routes such as WebSocket and tunnel paths.
type Controller struct {
	gatewaySvc gatewaysvc.Service
}

// New creates the AgentBox raw gateway controller.
func New(gatewaySvc gatewaysvc.Service) (*Controller, error) {
	if gatewaySvc == nil {
		return nil, gerror.New("agentbox gateway service is required")
	}
	return &Controller{gatewaySvc: gatewaySvc}, nil
}

// AgentShell validates shell WebSocket ownership before runtime handling.
func (c *Controller) AgentShell(r *ghttp.Request) {
	userID, err := authctx.RequireUserID(r.Context())
	if err != nil {
		r.SetError(err)
		return
	}
	err = c.gatewaySvc.AgentShell(
		r.Context(),
		userID,
		r.GetRouter("id").String(),
		r.Get("terminalId").String(),
		r.Get("cwd").String(),
		r.Get("mode").String(),
	)
	if err != nil {
		r.SetError(err)
	}
}

// AgentServiceHTTPProxy validates HTTP service proxy access before runtime handling.
func (c *Controller) AgentServiceHTTPProxy(r *ghttp.Request) {
	userID, err := authctx.RequireUserID(r.Context())
	if err != nil {
		r.SetError(err)
		return
	}
	if err := c.gatewaySvc.AgentServiceHTTPProxy(r.Context(), userID, r.URL.EscapedPath()); err != nil {
		r.SetError(err)
	}
}

// AgentChat validates Chat WebSocket ownership before runtime handling.
func (c *Controller) AgentChat(r *ghttp.Request) {
	userID, err := authctx.RequireUserID(r.Context())
	if err != nil {
		r.SetError(err)
		return
	}
	err = c.gatewaySvc.AgentChat(
		r.Context(),
		userID,
		r.GetRouter("id").String(),
		r.GetRouter("sessionId").String(),
		r.Get("cwd").String(),
	)
	if err != nil {
		r.SetError(err)
	}
}

// AgentServiceTCPTunnel validates TCP tunnel ownership before runtime handling.
func (c *Controller) AgentServiceTCPTunnel(r *ghttp.Request) {
	userID, err := authctx.RequireUserID(r.Context())
	if err != nil {
		r.SetError(err)
		return
	}
	err = c.gatewaySvc.AgentServiceTCPTunnel(
		r.Context(),
		userID,
		r.GetRouter("id").String(),
		r.GetRouter("serviceId").String(),
		r.Get("key").String(),
	)
	if err != nil {
		r.SetError(err)
	}
}
