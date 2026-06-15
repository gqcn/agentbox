// This file converts AgentBox catalog coding-image projections into versioned
// API DTOs without exposing service internals through controller methods.

package image

import (
	v1 "john-ai-agentbox/backend/api/image/v1"
	catalogsvc "john-ai-agentbox/backend/internal/service/catalog"
)

// toImageResponse maps a service coding-image projection into a public DTO.
func toImageResponse(item catalogsvc.CodingImageInfo) v1.CodingImageInfo {
	return v1.CodingImageInfo{
		ID:           item.ID,
		Name:         item.Name,
		ImageRef:     item.ImageRef,
		AgentType:    item.AgentType,
		DefaultShell: item.DefaultShell,
		Notes:        item.Notes,
		Enabled:      item.Enabled,
		IsDefault:    item.IsDefault,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	}
}

// toImageListResponse maps service image projections into public DTOs.
func toImageListResponse(items []catalogsvc.CodingImageInfo) []v1.CodingImageInfo {
	result := make([]v1.CodingImageInfo, 0, len(items))
	for _, item := range items {
		result = append(result, toImageResponse(item))
	}
	return result
}
