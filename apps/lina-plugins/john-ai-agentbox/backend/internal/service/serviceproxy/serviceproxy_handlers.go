// This file implements runtime service and bridge migration behavior. It
// deliberately checks AgentBox ownership before reporting runtime availability
// so invisible Agents and services do not leak proxy or tunnel metadata.

package serviceproxy

import (
	"context"
	"net"
	"strings"

	"lina-core/pkg/bizerr"
)

// Services lists runtime services for one visible Agent.
func (s *serviceImpl) Services(ctx context.Context, userID string, agentID string) ([]RuntimeServiceInfo, error) {
	userID, agentID = normalizeOwnerAndAgent(userID, agentID)
	if err := s.accessSvc.EnsureAgentVisible(ctx, userID, agentID); err != nil {
		return nil, err
	}
	if s.runtimeBackend == nil {
		return nil, bizerr.NewCode(CodeServiceProxyRuntimeUnavailable)
	}
	items, err := s.runtimeBackend.RuntimeServices(ctx, userID, agentID)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeServiceProxyRuntimeUnavailable)
	}
	if len(items) > s.config.RuntimeServiceListLimit {
		items = items[:s.config.RuntimeServiceListLimit]
	}
	return items, nil
}

// Service gets one runtime service for one visible Agent.
func (s *serviceImpl) Service(ctx context.Context, userID string, agentID string, serviceID string) (RuntimeServiceInfo, error) {
	userID, agentID, serviceID = normalizeServiceScope(userID, agentID, serviceID)
	if err := s.accessSvc.EnsureServiceProxyVisible(ctx, userID, agentID, serviceID); err != nil {
		return RuntimeServiceInfo{}, err
	}
	if s.runtimeBackend == nil {
		return RuntimeServiceInfo{}, bizerr.NewCode(CodeServiceProxyRuntimeUnavailable)
	}
	items, err := s.runtimeBackend.RuntimeServices(ctx, userID, agentID)
	if err != nil {
		return RuntimeServiceInfo{}, bizerr.WrapCode(err, CodeServiceProxyRuntimeUnavailable)
	}
	for _, item := range items {
		if item.ID == serviceID {
			return item, nil
		}
	}
	return RuntimeServiceInfo{}, bizerr.NewCode(CodeServiceProxyRuntimeUnavailable)
}

// ServiceBridges lists explicit bridges for one visible Agent.
func (s *serviceImpl) ServiceBridges(ctx context.Context, userID string, agentID string) ([]BridgeInfo, error) {
	userID, agentID = normalizeOwnerAndAgent(userID, agentID)
	if err := s.accessSvc.EnsureAgentVisible(ctx, userID, agentID); err != nil {
		return nil, err
	}
	return nil, bizerr.NewCode(CodeServiceProxyRuntimeUnavailable)
}

// CreateServiceBridge creates an explicit loopback bridge for one visible service.
func (s *serviceImpl) CreateServiceBridge(ctx context.Context, userID string, agentID string, input BridgeInput) (BridgeInfo, error) {
	userID, agentID, input.ServiceID = normalizeServiceScope(userID, agentID, input.ServiceID)
	input.ListenAddress = strings.TrimSpace(input.ListenAddress)
	if err := validateBridgeInput(input); err != nil {
		return BridgeInfo{}, err
	}
	if err := s.accessSvc.EnsureServiceProxyVisible(ctx, userID, agentID, input.ServiceID); err != nil {
		return BridgeInfo{}, err
	}
	return BridgeInfo{}, bizerr.NewCode(CodeServiceProxyRuntimeUnavailable)
}

// DeleteServiceBridge closes one explicit bridge for a visible Agent.
func (s *serviceImpl) DeleteServiceBridge(ctx context.Context, userID string, agentID string, bridgeID string) (bool, error) {
	userID, agentID = normalizeOwnerAndAgent(userID, agentID)
	bridgeID = strings.TrimSpace(bridgeID)
	if bridgeID == "" {
		return false, bizerr.NewCode(CodeServiceProxyInvalidInput)
	}
	if err := s.accessSvc.EnsureAgentVisible(ctx, userID, agentID); err != nil {
		return false, err
	}
	return false, bizerr.NewCode(CodeServiceProxyRuntimeUnavailable)
}

func validateBridgeInput(input BridgeInput) error {
	if input.ServiceID == "" || input.ListenAddress == "" || input.Port < 1 || input.Port > 65535 {
		return bizerr.NewCode(CodeServiceProxyInvalidInput)
	}
	ip := net.ParseIP(input.ListenAddress)
	if ip == nil || !ip.IsLoopback() {
		return bizerr.NewCode(CodeServiceProxyInvalidInput)
	}
	return nil
}

func normalizeOwnerAndAgent(userID string, agentID string) (string, string) {
	return strings.TrimSpace(userID), strings.TrimSpace(agentID)
}

func normalizeServiceScope(userID string, agentID string, serviceID string) (string, string, string) {
	return strings.TrimSpace(userID), strings.TrimSpace(agentID), strings.TrimSpace(serviceID)
}
