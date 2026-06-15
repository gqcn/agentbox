// This file maps AgentBox service-proxy service projections to public DTOs
// while keeping runtime backend internals out of HTTP response contracts.

package serviceproxy

import (
	v1 "john-ai-agentbox/backend/api/serviceproxy/v1"
	serviceproxysvc "john-ai-agentbox/backend/internal/service/serviceproxy"
)

func toRuntimeServiceListResponse(items []serviceproxysvc.RuntimeServiceInfo) []v1.AgentRuntimeServiceInfo {
	out := make([]v1.AgentRuntimeServiceInfo, 0, len(items))
	for _, item := range items {
		out = append(out, toRuntimeServiceResponse(item))
	}
	return out
}

func toRuntimeServiceResponse(item serviceproxysvc.RuntimeServiceInfo) v1.AgentRuntimeServiceInfo {
	return v1.AgentRuntimeServiceInfo{
		ID:                item.ID,
		AgentID:           item.AgentID,
		Port:              item.Port,
		Protocol:          item.Protocol,
		AccessStatus:      item.AccessStatus,
		ListenAddresses:   toListenAddressListResponse(item.ListenAddresses),
		ProcessName:       item.ProcessName,
		PID:               item.PID,
		ProxyURL:          item.ProxyURL,
		TunnelURL:         item.TunnelURL,
		TunnelCommand:     item.TunnelCommand,
		BridgeID:          item.BridgeID,
		LocalHost:         item.LocalHost,
		LocalPort:         item.LocalPort,
		UnavailableReason: item.UnavailableReason,
		LastCheckedAt:     item.LastCheckedAt,
	}
}

func toListenAddressListResponse(items []serviceproxysvc.ListenAddress) []v1.AgentServiceListenAddress {
	out := make([]v1.AgentServiceListenAddress, 0, len(items))
	for _, item := range items {
		out = append(out, v1.AgentServiceListenAddress{
			Address:           item.Address,
			Port:              item.Port,
			Network:           item.Network,
			AccessStatus:      item.AccessStatus,
			BridgeID:          item.BridgeID,
			LocalHost:         item.LocalHost,
			LocalPort:         item.LocalPort,
			UnavailableReason: item.UnavailableReason,
		})
	}
	return out
}

func toBridgeListResponse(items []serviceproxysvc.BridgeInfo) []v1.AgentServiceBridgeInfo {
	out := make([]v1.AgentServiceBridgeInfo, 0, len(items))
	for _, item := range items {
		out = append(out, toBridgeResponse(item))
	}
	return out
}

func toBridgeResponse(item serviceproxysvc.BridgeInfo) v1.AgentServiceBridgeInfo {
	return v1.AgentServiceBridgeInfo{
		ID:            item.ID,
		AgentID:       item.AgentID,
		ServiceID:     item.ServiceID,
		ListenAddress: item.ListenAddress,
		Port:          item.Port,
		BridgePort:    item.BridgePort,
		LocalHost:     item.LocalHost,
		LocalPort:     item.LocalPort,
		Status:        item.Status,
		ErrorMessage:  item.ErrorMessage,
		CreatedAt:     item.CreatedAt,
		ClosedAt:      item.ClosedAt,
	}
}
