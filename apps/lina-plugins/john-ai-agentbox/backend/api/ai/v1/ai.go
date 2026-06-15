// This file defines version-one AI capability DTOs for the AgentBox plugin.
// Paths are plugin-relative and are published under /x/john-ai-agentbox/api/v1
// by source-plugin route registration.

package v1

import "github.com/gogf/gf/v2/frame/g"

// CapabilityTiersReq lists fixed AgentBox AI capability tiers.
type CapabilityTiersReq struct {
	g.Meta `path:"/ai/capability-tiers" method:"get" tags:"AgentBox AI" summary:"List AgentBox AI capability tiers" dc:"List basic, standard, and advanced AgentBox AI capability tiers with the current primary binding and latest test status."`
}

// CapabilityTiersRes returns AI capability tiers.
type CapabilityTiersRes = []AICapabilityTierInfo

// UpdateCapabilityTierReq updates one AI capability tier.
type UpdateCapabilityTierReq struct {
	g.Meta          `path:"/ai/capability-tiers/{code}" method:"put" tags:"AgentBox AI" summary:"Update AgentBox AI capability tier" dc:"Update whether one AgentBox AI capability tier is enabled and optionally set its primary provider model binding."`
	Code            string `json:"code" v:"required|in:basic,standard,advanced" dc:"AI capability tier code: basic, standard, advanced" eg:"basic"`
	Enabled         bool   `json:"enabled" dc:"Whether this AI capability tier can be used by AgentBox AI calls" eg:"true"`
	ProviderID      int64  `json:"providerId" dc:"Provider ID for the primary tier binding; 0 means keep existing binding" eg:"1"`
	ProviderModelID int64  `json:"providerModelId" dc:"Provider model ID for the primary tier binding; 0 means keep existing binding" eg:"10"`
	Protocol        string `json:"protocol" dc:"Optional target provider protocol for the binding: openai or anthropic; omitted means using the selected model record protocol" eg:"openai"`
}

// UpdateCapabilityTierRes returns the updated tier.
type UpdateCapabilityTierRes = AICapabilityTierInfo

// TestCapabilityTierReq tests one AI capability tier or a draft provider/model.
type TestCapabilityTierReq struct {
	g.Meta          `path:"/ai/capability-tiers/{code}/test" method:"post" tags:"AgentBox AI" summary:"Test AgentBox AI capability tier" dc:"Run a lightweight text generation test for one AgentBox AI capability tier. Draft provider and model values are tested without changing the saved tier binding."`
	Code            string `json:"code" v:"required|in:basic,standard,advanced" dc:"AI capability tier code: basic, standard, advanced" eg:"basic"`
	ProviderID      int64  `json:"providerId" dc:"Optional draft provider ID to test without saving; 0 means test the saved tier binding" eg:"1"`
	ProviderModelID int64  `json:"providerModelId" dc:"Optional draft provider model ID to test without saving; 0 means test the saved tier binding" eg:"10"`
	Protocol        string `json:"protocol" dc:"Optional target protocol for the draft binding: openai or anthropic; omitted means using the selected model record protocol" eg:"openai"`
}

// TestCapabilityTierRes returns tier test details.
type TestCapabilityTierRes = AICapabilityTestResult

// InvocationsReq lists AI invocation logs.
type InvocationsReq struct {
	g.Meta  `path:"/ai/invocations" method:"get" tags:"AgentBox AI" summary:"List AgentBox AI invocations" dc:"List sanitized AgentBox AI invocation audit records ordered by creation time descending. Prompt text, Git diffs, and full provider responses are never returned."`
	Purpose string `json:"purpose" dc:"Optional purpose filter: capability_test or git_commit_message; omitted means all purposes" eg:"git_commit_message"`
	Tier    string `json:"tier" dc:"Optional tier filter: basic, standard, advanced; omitted means all tiers" eg:"basic"`
	Status  string `json:"status" dc:"Optional invocation status filter: success or error; omitted means all statuses" eg:"success"`
	Limit   int    `json:"limit" dc:"Maximum rows to return, default 50 and maximum 200" eg:"50"`
}

// InvocationsRes returns invocation log records.
type InvocationsRes = []AIInvocationLogInfo

// AICapabilityTierInfo describes one fixed AgentBox AI capability tier.
type AICapabilityTierInfo struct {
	Code        string                   `json:"code" dc:"AI capability tier code: basic, standard, advanced" eg:"basic"`
	DisplayName string                   `json:"displayName" dc:"Human-readable tier name" eg:"基础"`
	Description string                   `json:"description" dc:"Tier usage description" eg:"基础 AI 能力，用于 Git commit message、标题摘要和简单文本生成。"`
	Enabled     bool                     `json:"enabled" dc:"Whether this tier is enabled for AgentBox AI calls" eg:"true"`
	Configured  bool                     `json:"configured" dc:"Whether this tier has an enabled primary provider model binding" eg:"true"`
	Available   bool                     `json:"available" dc:"Whether this tier is enabled and configured for AgentBox AI calls" eg:"true"`
	Binding     *AICapabilityBindingInfo `json:"binding,omitempty" dc:"Primary provider model binding for this tier, omitted when unconfigured" eg:"{}"`
	LastTest    *AIInvocationLogInfo     `json:"lastTest,omitempty" dc:"Latest lightweight capability test invocation log, omitted when never tested" eg:"{}"`
	CreatedAt   int64                    `json:"createdAt" dc:"Tier creation time as Unix timestamp in milliseconds" eg:"1704067200000"`
	UpdatedAt   int64                    `json:"updatedAt" dc:"Tier update time as Unix timestamp in milliseconds" eg:"1704067200000"`
}

// AICapabilityBindingInfo describes a tier-to-provider-model binding.
type AICapabilityBindingInfo struct {
	ID              int64  `json:"id" dc:"AI capability binding ID" eg:"1"`
	TierCode        string `json:"tierCode" dc:"AI capability tier code for this binding: basic, standard, advanced" eg:"basic"`
	ProviderID      int64  `json:"providerId" dc:"Bound AI provider ID" eg:"1"`
	ProviderName    string `json:"providerName" dc:"Bound AI provider display name" eg:"OpenAI"`
	ProviderModelID int64  `json:"providerModelId" dc:"Bound provider model ID" eg:"10"`
	ModelName       string `json:"modelName" dc:"Bound provider model name" eg:"gpt-5-codex"`
	Protocol        string `json:"protocol" dc:"Provider model protocol: openai or anthropic" eg:"openai"`
	Priority        int    `json:"priority" dc:"Binding priority, 0 is the primary binding" eg:"0"`
	Enabled         bool   `json:"enabled" dc:"Whether this binding can be used by tier resolution" eg:"true"`
	CreatedAt       int64  `json:"createdAt" dc:"Binding creation time as Unix timestamp in milliseconds" eg:"1704067200000"`
	UpdatedAt       int64  `json:"updatedAt" dc:"Binding update time as Unix timestamp in milliseconds" eg:"1704067200000"`
}

// AIInvocationLogInfo exposes the minimal audit record for one AgentBox AI call.
type AIInvocationLogInfo struct {
	ID              int64  `json:"id" dc:"AI invocation log ID" eg:"1"`
	Purpose         string `json:"purpose" dc:"AI invocation purpose: capability_test or git_commit_message" eg:"git_commit_message"`
	TierCode        string `json:"tierCode" dc:"AI capability tier code used by the invocation: basic, standard, advanced" eg:"basic"`
	ProviderID      int64  `json:"providerId,omitempty" dc:"Actual provider ID used by the invocation, omitted when unavailable" eg:"1"`
	ProviderName    string `json:"providerName,omitempty" dc:"Actual provider display name used by the invocation, omitted when unavailable" eg:"OpenAI"`
	ProviderModelID int64  `json:"providerModelId,omitempty" dc:"Actual provider model ID used by the invocation, omitted when unavailable" eg:"10"`
	ModelName       string `json:"modelName,omitempty" dc:"Actual model name used by the invocation" eg:"gpt-5-codex"`
	Protocol        string `json:"protocol,omitempty" dc:"Actual provider protocol: openai or anthropic" eg:"openai"`
	Status          string `json:"status" dc:"Invocation status: success or error" eg:"success"`
	LatencyMS       int64  `json:"latencyMs" dc:"Invocation latency in milliseconds" eg:"328"`
	ErrorMessage    string `json:"errorMessage,omitempty" dc:"Sanitized error summary for failed invocations" eg:"ai provider request failed"`
	CreatedAt       int64  `json:"createdAt" dc:"Invocation creation time as Unix timestamp in milliseconds" eg:"1704067200000"`
}

// AICapabilityTestResult returns the result of a lightweight tier test.
type AICapabilityTestResult struct {
	Status          string `json:"status" dc:"Capability test status: success or error" eg:"success"`
	TierCode        string `json:"tierCode" dc:"AI capability tier code tested: basic, standard, advanced" eg:"basic"`
	ProviderID      int64  `json:"providerId,omitempty" dc:"Provider ID used by the test, omitted when unavailable" eg:"1"`
	ProviderName    string `json:"providerName,omitempty" dc:"Provider display name used by the test, omitted when unavailable" eg:"OpenAI"`
	ProviderModelID int64  `json:"providerModelId,omitempty" dc:"Provider model ID used by the test, omitted when unavailable" eg:"10"`
	ModelName       string `json:"modelName,omitempty" dc:"Model name used by the test" eg:"gpt-5-codex"`
	Protocol        string `json:"protocol,omitempty" dc:"Provider protocol used by the test: openai or anthropic" eg:"openai"`
	LatencyMS       int64  `json:"latencyMs" dc:"Capability test latency in milliseconds" eg:"328"`
	ErrorMessage    string `json:"errorMessage,omitempty" dc:"Sanitized error summary when the capability test fails" eg:"ai provider request failed"`
	TestedAt        int64  `json:"testedAt" dc:"Capability test time as Unix timestamp in milliseconds" eg:"1704067200000"`
}
