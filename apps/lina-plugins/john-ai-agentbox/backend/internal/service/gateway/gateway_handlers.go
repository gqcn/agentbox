// This file implements the current raw gateway migration behavior. It checks
// ownership before reporting runtime availability so WebSocket and tunnel paths
// cannot leak other users' Agent, Chat, terminal, or service identifiers.

package gateway

import (
	"context"
	"net/url"
	"strings"

	"lina-core/pkg/bizerr"
)

const serviceProxyPathMarker = "/proxy/"

// AgentServiceHTTPProxy validates the raw HTTP service proxy path before runtime handling.
func (s *serviceImpl) AgentServiceHTTPProxy(ctx context.Context, userID string, escapedPath string) error {
	_ = ctx
	if strings.TrimSpace(userID) == "" {
		return bizerr.NewCode(CodeGatewayInvalidInput)
	}
	if _, _, err := parseServiceProxyPath(escapedPath); err != nil {
		return err
	}
	return bizerr.NewCode(CodeGatewayRuntimeUnavailable)
}

// AgentShell validates Agent ownership before opening a runtime-backed shell WebSocket.
func (s *serviceImpl) AgentShell(ctx context.Context, userID string, agentID string, terminalID string, cwd string, mode string) error {
	_ = cwd
	_ = mode
	userID, agentID = strings.TrimSpace(userID), strings.TrimSpace(agentID)
	terminalID = strings.TrimSpace(terminalID)
	var err error
	if terminalID != "" {
		err = s.accessSvc.EnsureTerminalSessionVisible(ctx, userID, agentID, terminalID)
	} else {
		err = s.accessSvc.EnsureAgentVisible(ctx, userID, agentID)
	}
	if err != nil {
		return err
	}
	return bizerr.NewCode(CodeGatewayRuntimeUnavailable)
}

func parseServiceProxyPath(escapedPath string) (string, string, error) {
	index := strings.Index(escapedPath, serviceProxyPathMarker)
	if index < 0 {
		return "", "", bizerr.NewCode(CodeGatewayInvalidInput)
	}
	trimmed := strings.TrimSpace(escapedPath[index+len(serviceProxyPathMarker):])
	if trimmed == "" {
		return "", "", bizerr.NewCode(CodeGatewayInvalidInput)
	}
	parts := strings.SplitN(trimmed, "/", 2)
	key, err := url.PathUnescape(parts[0])
	if err != nil {
		return "", "", bizerr.WrapCode(err, CodeGatewayInvalidInput)
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return "", "", bizerr.NewCode(CodeGatewayInvalidInput)
	}
	restPath := ""
	if len(parts) == 2 {
		restPath, err = url.PathUnescape(parts[1])
		if err != nil {
			return "", "", bizerr.WrapCode(err, CodeGatewayInvalidInput)
		}
	}
	return key, restPath, nil
}

// AgentChat validates Agent Chat session ownership before opening a runtime-backed Chat WebSocket.
func (s *serviceImpl) AgentChat(ctx context.Context, userID string, agentID string, sessionID string, cwd string) error {
	_ = cwd
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return bizerr.NewCode(CodeGatewayInvalidInput)
	}
	if err := s.accessSvc.EnsureChatSessionVisible(ctx, userID, agentID, sessionID); err != nil {
		return err
	}
	return bizerr.NewCode(CodeGatewayRuntimeUnavailable)
}

// AgentServiceTCPTunnel validates service ownership before opening a runtime-backed TCP tunnel.
func (s *serviceImpl) AgentServiceTCPTunnel(ctx context.Context, userID string, agentID string, serviceID string, key string) error {
	_ = key
	serviceID = strings.TrimSpace(serviceID)
	if serviceID == "" {
		return bizerr.NewCode(CodeGatewayInvalidInput)
	}
	if err := s.accessSvc.EnsureServiceProxyVisible(ctx, userID, agentID, serviceID); err != nil {
		return err
	}
	return bizerr.NewCode(CodeGatewayRuntimeUnavailable)
}
