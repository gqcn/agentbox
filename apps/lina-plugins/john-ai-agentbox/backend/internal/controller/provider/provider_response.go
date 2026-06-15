// This file converts AgentBox catalog provider projections into versioned API
// DTOs without leaking service internals into controller methods.

package provider

import (
	v1 "john-ai-agentbox/backend/api/provider/v1"
	catalogsvc "john-ai-agentbox/backend/internal/service/catalog"
)

// toProviderResponse maps a service provider projection into a public DTO.
func toProviderResponse(item catalogsvc.ProviderInfo) v1.ProviderInfo {
	models := make([]v1.ProviderModelInfo, 0, len(item.Models))
	for _, model := range item.Models {
		models = append(models, toProviderModelResponse(model))
	}
	return v1.ProviderInfo{
		ID:               item.ID,
		Name:             item.Name,
		HomepageURL:      item.HomepageURL,
		Notes:            item.Notes,
		APIKeyMasked:     item.APIKeyMasked,
		APIKeyConfigured: item.APIKeyConfigured,
		OpenAIBaseURL:    item.OpenAIBaseURL,
		AnthropicBaseURL: item.AnthropicBaseURL,
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
		Models:           models,
	}
}

// toProviderModelResponse maps a service model projection into a public DTO.
func toProviderModelResponse(item catalogsvc.ProviderModelInfo) v1.ProviderModelInfo {
	return v1.ProviderModelInfo{
		ID:           item.ID,
		ProviderID:   item.ProviderID,
		Name:         item.Name,
		Protocol:     item.Protocol,
		Source:       item.Source,
		LastSyncedAt: item.LastSyncedAt,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	}
}

// toProviderListResponse maps service provider projections into public DTOs.
func toProviderListResponse(items []catalogsvc.ProviderInfo) []v1.ProviderInfo {
	result := make([]v1.ProviderInfo, 0, len(items))
	for _, item := range items {
		result = append(result, toProviderResponse(item))
	}
	return result
}

// toSyncModelsResponse maps a service sync result into a public DTO.
func toSyncModelsResponse(out *catalogsvc.SyncProviderModelsOutput) *v1.SyncProviderModelsResponse {
	if out == nil {
		return &v1.SyncProviderModelsResponse{}
	}
	models := make([]v1.ProviderModelInfo, 0, len(out.Models))
	for _, item := range out.Models {
		models = append(models, toProviderModelResponse(item))
	}
	return &v1.SyncProviderModelsResponse{
		Protocol: out.Protocol,
		Count:    out.Count,
		Models:   models,
	}
}
