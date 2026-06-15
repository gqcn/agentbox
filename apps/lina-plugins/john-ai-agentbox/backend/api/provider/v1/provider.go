// This file defines version-one provider DTOs for the AgentBox plugin. Paths
// are plugin-relative and are published under /x/john-ai-agentbox/api/v1 by
// source-plugin route registration.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListReq lists configured AI providers.
type ListReq struct {
	g.Meta `path:"/providers" method:"get" tags:"AgentBox Providers" summary:"List AgentBox providers" dc:"List configured AgentBox AI providers together with their available model records."`
}

// ListRes returns configured AI providers.
type ListRes = []ProviderInfo

// CreateReq creates an AI provider.
type CreateReq struct {
	g.Meta           `path:"/providers" method:"post" tags:"AgentBox Providers" summary:"Create AgentBox provider" dc:"Create an AgentBox AI provider configuration used by coding agents to connect to OpenAI-compatible or Anthropic-compatible APIs."`
	Name             string `json:"name" v:"required" dc:"Provider display name" eg:"OpenAI"`
	HomepageURL      string `json:"homepageUrl" dc:"Provider homepage URL" eg:"https://openai.com"`
	Notes            string `json:"notes" dc:"Operator notes for this provider" eg:"Primary OpenAI-compatible endpoint"`
	APIKey           string `json:"apiKey" dc:"Provider API key; stored server-side and returned only as a masked value" eg:"sk-xxxx"`
	OpenAIBaseURL    string `json:"openaiBaseUrl" dc:"OpenAI-compatible base URL" eg:"https://api.openai.com/v1"`
	AnthropicBaseURL string `json:"anthropicBaseUrl" dc:"Anthropic-compatible base URL" eg:"https://api.anthropic.com"`
}

// CreateRes returns the created provider.
type CreateRes = ProviderInfo

// DetailReq gets one provider.
type DetailReq struct {
	g.Meta `path:"/providers/{id}" method:"get" tags:"AgentBox Providers" summary:"Get AgentBox provider" dc:"Get one AgentBox AI provider and its available model records."`
	ID     int64 `json:"id" v:"required|min:1" dc:"Provider ID" eg:"1"`
}

// DetailRes returns one provider.
type DetailRes = ProviderInfo

// UpdateReq updates one provider.
type UpdateReq struct {
	g.Meta           `path:"/providers/{id}" method:"put" tags:"AgentBox Providers" summary:"Update AgentBox provider" dc:"Update an AgentBox AI provider configuration; an empty API key keeps the existing secret unchanged."`
	ID               int64  `json:"id" v:"required|min:1" dc:"Provider ID" eg:"1"`
	Name             string `json:"name" v:"required" dc:"Provider display name" eg:"OpenAI"`
	HomepageURL      string `json:"homepageUrl" dc:"Provider homepage URL" eg:"https://openai.com"`
	Notes            string `json:"notes" dc:"Operator notes for this provider" eg:"Primary OpenAI-compatible endpoint"`
	APIKey           string `json:"apiKey" dc:"Provider API key; omitted or empty means keeping the existing secret" eg:"sk-xxxx"`
	OpenAIBaseURL    string `json:"openaiBaseUrl" dc:"OpenAI-compatible base URL" eg:"https://api.openai.com/v1"`
	AnthropicBaseURL string `json:"anthropicBaseUrl" dc:"Anthropic-compatible base URL" eg:"https://api.anthropic.com"`
}

// UpdateRes returns the updated provider.
type UpdateRes = ProviderInfo

// DeleteReq deletes one provider.
type DeleteReq struct {
	g.Meta `path:"/providers/{id}" method:"delete" tags:"AgentBox Providers" summary:"Delete AgentBox provider" dc:"Delete an AgentBox AI provider if it is not referenced by agents, provider model records, or AI capability bindings."`
	ID     int64 `json:"id" v:"required|min:1" dc:"Provider ID" eg:"1"`
}

// DeleteRes reports deletion state.
type DeleteRes struct {
	Deleted bool `json:"deleted" dc:"Whether the provider was deleted" eg:"true"`
}

// CreateModelReq creates one provider model.
type CreateModelReq struct {
	g.Meta   `path:"/providers/{id}/models" method:"post" tags:"AgentBox Providers" summary:"Create AgentBox provider model" dc:"Create a manually managed model record for the provider."`
	ID       int64  `json:"id" v:"required|min:1" dc:"Provider ID" eg:"1"`
	Name     string `json:"name" v:"required" dc:"Model name exposed by the provider" eg:"gpt-5"`
	Protocol string `json:"protocol" v:"required|in:openai,anthropic" dc:"Model protocol: openai=OpenAI-compatible, anthropic=Anthropic-compatible" eg:"openai"`
}

// CreateModelRes returns the created provider model.
type CreateModelRes = ProviderModelInfo

// DeleteModelReq deletes one provider model.
type DeleteModelReq struct {
	g.Meta  `path:"/providers/{id}/models/{modelId}" method:"delete" tags:"AgentBox Providers" summary:"Delete AgentBox provider model" dc:"Delete one provider model record if no agent or AI capability binding references it."`
	ID      int64 `json:"id" v:"required|min:1" dc:"Provider ID" eg:"1"`
	ModelID int64 `json:"modelId" v:"required|min:1" dc:"Provider model ID" eg:"10"`
}

// DeleteModelRes reports provider model deletion state.
type DeleteModelRes struct {
	Deleted bool `json:"deleted" dc:"Whether the provider model was deleted" eg:"true"`
}

// SyncModelsReq synchronizes provider models from the remote provider API.
type SyncModelsReq struct {
	g.Meta   `path:"/providers/{id}/models/sync" method:"post" tags:"AgentBox Providers" summary:"Sync AgentBox provider models" dc:"Fetch the remote model list for the selected protocol and upsert provider model records with a bounded synchronization limit."`
	ID       int64  `json:"id" v:"required|min:1" dc:"Provider ID" eg:"1"`
	Protocol string `json:"protocol" v:"required|in:openai,anthropic" dc:"Model protocol to synchronize: openai=OpenAI-compatible, anthropic=Anthropic-compatible" eg:"openai"`
}

// SyncModelsRes returns synchronized provider model details.
type SyncModelsRes = SyncProviderModelsResponse

// ProviderInfo is the public AgentBox provider projection.
type ProviderInfo struct {
	ID               int64               `json:"id" dc:"Provider ID" eg:"1"`
	Name             string              `json:"name" dc:"Provider display name" eg:"OpenAI"`
	HomepageURL      string              `json:"homepageUrl" dc:"Provider homepage URL" eg:"https://openai.com"`
	Notes            string              `json:"notes" dc:"Operator notes for this provider" eg:"Primary OpenAI-compatible endpoint"`
	APIKeyMasked     string              `json:"apiKeyMasked" dc:"Masked provider API key; empty when no key is configured" eg:"sk-****abcd"`
	APIKeyConfigured bool                `json:"apiKeyConfigured" dc:"Whether a provider API key is configured" eg:"true"`
	OpenAIBaseURL    string              `json:"openaiBaseUrl" dc:"OpenAI-compatible base URL" eg:"https://api.openai.com/v1"`
	AnthropicBaseURL string              `json:"anthropicBaseUrl" dc:"Anthropic-compatible base URL" eg:"https://api.anthropic.com"`
	CreatedAt        int64               `json:"createdAt" dc:"Creation time as Unix timestamp in milliseconds" eg:"1704067200000"`
	UpdatedAt        int64               `json:"updatedAt" dc:"Last update time as Unix timestamp in milliseconds" eg:"1704067201000"`
	Models           []ProviderModelInfo `json:"models,omitempty" dc:"Provider model records available for agent configuration" eg:"[]"`
}

// ProviderModelInfo is the public AgentBox provider model projection.
type ProviderModelInfo struct {
	ID           int64  `json:"id" dc:"Provider model ID" eg:"10"`
	ProviderID   int64  `json:"providerId" dc:"Owning provider ID" eg:"1"`
	Name         string `json:"name" dc:"Model name exposed by the provider" eg:"gpt-5"`
	Protocol     string `json:"protocol" dc:"Model protocol: openai=OpenAI-compatible, anthropic=Anthropic-compatible" eg:"openai"`
	Source       string `json:"source" dc:"Model source: manual=operator managed, api=synchronized from provider API" eg:"api"`
	LastSyncedAt *int64 `json:"lastSyncedAt,omitempty" dc:"Last provider API synchronization time as Unix timestamp in milliseconds; omitted for manually maintained models" eg:"1704067200000"`
	CreatedAt    int64  `json:"createdAt" dc:"Creation time as Unix timestamp in milliseconds" eg:"1704067200000"`
	UpdatedAt    int64  `json:"updatedAt" dc:"Last update time as Unix timestamp in milliseconds" eg:"1704067201000"`
}

// SyncProviderModelsResponse reports the bounded remote model sync result.
type SyncProviderModelsResponse struct {
	Protocol string              `json:"protocol" dc:"Model protocol synchronized: openai=OpenAI-compatible, anthropic=Anthropic-compatible" eg:"openai"`
	Count    int                 `json:"count" dc:"Number of provider model records synchronized" eg:"3"`
	Models   []ProviderModelInfo `json:"models" dc:"Synchronized provider model records" eg:"[]"`
}
