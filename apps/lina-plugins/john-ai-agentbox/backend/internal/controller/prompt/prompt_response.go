// This file maps AgentBox prompt service projections to public DTOs. It keeps
// service-internal registry structures out of HTTP responses.

package prompt

import (
	v1 "john-ai-agentbox/backend/api/prompt/v1"
	promptsvc "john-ai-agentbox/backend/internal/service/prompt"
)

func toPromptListResponse(items []promptsvc.TemplateInfo) []v1.PromptTemplateInfo {
	out := make([]v1.PromptTemplateInfo, 0, len(items))
	for _, item := range items {
		out = append(out, toPromptResponse(item))
	}
	return out
}

func toPromptResponse(item promptsvc.TemplateInfo) v1.PromptTemplateInfo {
	return v1.PromptTemplateInfo{
		Code:           item.Code,
		DisplayName:    item.DisplayName,
		Description:    item.Description,
		Purpose:        item.Purpose,
		TierCode:       item.TierCode,
		DefaultContent: item.DefaultContent,
		Content:        item.Content,
		Variables:      toPromptVariableListResponse(item.Variables),
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.UpdatedAt,
	}
}

func toPromptVariableListResponse(items []promptsvc.VariableInfo) []v1.PromptTemplateVariableInfo {
	out := make([]v1.PromptTemplateVariableInfo, 0, len(items))
	for _, item := range items {
		out = append(out, v1.PromptTemplateVariableInfo{
			Name:        item.Name,
			Description: item.Description,
			Required:    item.Required,
			SampleValue: item.SampleValue,
		})
	}
	return out
}
