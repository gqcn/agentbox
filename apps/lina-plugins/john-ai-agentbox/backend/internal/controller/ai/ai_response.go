// This file converts AgentBox AI service projections into versioned API DTOs
// without leaking service internals into controller methods.

package ai

import (
	v1 "john-ai-agentbox/backend/api/ai/v1"
	aisvc "john-ai-agentbox/backend/internal/service/ai"
)

func toTierListResponse(items []aisvc.CapabilityTierInfo) []v1.AICapabilityTierInfo {
	out := make([]v1.AICapabilityTierInfo, 0, len(items))
	for _, item := range items {
		out = append(out, toTierResponse(item))
	}
	return out
}

func toTierResponse(item aisvc.CapabilityTierInfo) v1.AICapabilityTierInfo {
	return v1.AICapabilityTierInfo{
		Code:        item.Code,
		DisplayName: item.DisplayName,
		Description: item.Description,
		Enabled:     item.Enabled,
		Configured:  item.Configured,
		Available:   item.Available,
		Binding:     toBindingResponse(item.Binding),
		LastTest:    toInvocationPtrResponse(item.LastTest),
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func toBindingResponse(item *aisvc.CapabilityBindingInfo) *v1.AICapabilityBindingInfo {
	if item == nil {
		return nil
	}
	return &v1.AICapabilityBindingInfo{
		ID:              item.ID,
		TierCode:        item.TierCode,
		ProviderID:      item.ProviderID,
		ProviderName:    item.ProviderName,
		ProviderModelID: item.ProviderModelID,
		ModelName:       item.ModelName,
		Protocol:        item.Protocol,
		Priority:        item.Priority,
		Enabled:         item.Enabled,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
	}
}

func toInvocationListResponse(items []aisvc.InvocationLogInfo) []v1.AIInvocationLogInfo {
	out := make([]v1.AIInvocationLogInfo, 0, len(items))
	for _, item := range items {
		out = append(out, toInvocationResponse(item))
	}
	return out
}

func toInvocationPtrResponse(item *aisvc.InvocationLogInfo) *v1.AIInvocationLogInfo {
	if item == nil {
		return nil
	}
	out := toInvocationResponse(*item)
	return &out
}

func toInvocationResponse(item aisvc.InvocationLogInfo) v1.AIInvocationLogInfo {
	return v1.AIInvocationLogInfo{
		ID:              item.ID,
		Purpose:         item.Purpose,
		TierCode:        item.TierCode,
		ProviderID:      item.ProviderID,
		ProviderName:    item.ProviderName,
		ProviderModelID: item.ProviderModelID,
		ModelName:       item.ModelName,
		Protocol:        item.Protocol,
		Status:          item.Status,
		LatencyMS:       item.LatencyMS,
		ErrorMessage:    item.ErrorMessage,
		CreatedAt:       item.CreatedAt,
	}
}

func toTestResponse(item aisvc.CapabilityTestResult) v1.AICapabilityTestResult {
	return v1.AICapabilityTestResult{
		Status:          item.Status,
		TierCode:        item.TierCode,
		ProviderID:      item.ProviderID,
		ProviderName:    item.ProviderName,
		ProviderModelID: item.ProviderModelID,
		ModelName:       item.ModelName,
		Protocol:        item.Protocol,
		LatencyMS:       item.LatencyMS,
		ErrorMessage:    item.ErrorMessage,
		TestedAt:        item.TestedAt,
	}
}
