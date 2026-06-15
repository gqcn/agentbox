// This file defines version-one coding-agent DTOs for the AgentBox plugin.
// Paths are plugin-relative and are published under /x/john-ai-agentbox/api/v1
// by source-plugin route registration.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListReq lists coding agents visible to the current AgentBox user.
type ListReq struct {
	g.Meta `path:"/agents" method:"get" tags:"AgentBox Agents" summary:"List AgentBox agents" dc:"List non-deleted coding agents owned by the authenticated AgentBox user with provider, image, and runtime projections."`
}

// ListRes returns coding agents.
type ListRes = []AgentInfo

// CreateReq creates one coding agent.
type CreateReq struct {
	g.Meta        `path:"/agents" method:"post" tags:"AgentBox Agents" summary:"Create AgentBox agent" dc:"Create a coding agent owned by the authenticated AgentBox user by binding a provider model and coding image profile."`
	Name          string `json:"name" v:"required" dc:"Agent display name" eg:"Frontend Workbench"`
	ProviderID    int64  `json:"providerId" v:"required|min:1" dc:"Provider ID used by this agent" eg:"1"`
	ModelName     string `json:"modelName" v:"required" dc:"Provider model name" eg:"gpt-5"`
	ModelProtocol string `json:"modelProtocol" v:"required|in:openai,anthropic" dc:"Model protocol: openai=OpenAI-compatible, anthropic=Anthropic-compatible" eg:"openai"`
	ImageID       int64  `json:"imageId" v:"required|min:1" dc:"Coding image ID used by this agent" eg:"1"`
	AgentType     string `json:"agentType" v:"required" dc:"Abstract agent type: claude_code, codex, custom" eg:"codex"`
	IconKey       string `json:"iconKey" dc:"Optional preset sidebar icon key; empty means using the default icon resolved from the abstract agent type" eg:"test-tube"`
	Notes         string `json:"notes" dc:"Operator notes for this agent" eg:"Primary frontend coding agent"`
}

// CreateRes returns the created coding agent.
type CreateRes = AgentInfo

// DetailReq gets one coding agent.
type DetailReq struct {
	g.Meta `path:"/agents/{id}" method:"get" tags:"AgentBox Agents" summary:"Get AgentBox agent" dc:"Get one coding agent owned by the authenticated AgentBox user with provider, image, and runtime projections."`
	ID     string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
}

// DetailRes returns one coding agent.
type DetailRes = AgentInfo

// UpdateReq updates one coding agent.
type UpdateReq struct {
	g.Meta        `path:"/agents/{id}" method:"put" tags:"AgentBox Agents" summary:"Update AgentBox agent" dc:"Update one coding agent owned by the authenticated AgentBox user's provider model binding, abstract agent type, and display metadata."`
	ID            string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	Name          string `json:"name" v:"required" dc:"Agent display name" eg:"Frontend Workbench"`
	ProviderID    int64  `json:"providerId" v:"required|min:1" dc:"Provider ID used by this agent" eg:"1"`
	ModelName     string `json:"modelName" v:"required" dc:"Provider model name" eg:"gpt-5"`
	ModelProtocol string `json:"modelProtocol" v:"required|in:openai,anthropic" dc:"Model protocol: openai=OpenAI-compatible, anthropic=Anthropic-compatible" eg:"openai"`
	AgentType     string `json:"agentType" v:"required" dc:"Abstract agent type: claude_code, codex, custom" eg:"codex"`
	IconKey       string `json:"iconKey" dc:"Optional preset sidebar icon key; empty means using the default icon resolved from the abstract agent type" eg:"git-pull-request"`
	Notes         string `json:"notes" dc:"Operator notes for this agent" eg:"Primary frontend coding agent"`
}

// UpdateRes returns the updated coding agent.
type UpdateRes = AgentInfo

// ChangeImageReq changes one agent image.
type ChangeImageReq struct {
	g.Meta  `path:"/agents/{id}/image" method:"put" tags:"AgentBox Agents" summary:"Change AgentBox agent image" dc:"Switch one authenticated-user-owned coding agent to a new image profile and report managed paths that may be reset or preserved."`
	ID      string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	ImageID int64  `json:"imageId" v:"required|min:1" dc:"New coding image ID" eg:"2"`
}

// ChangeImageRes returns image switch details.
type ChangeImageRes = ChangeAgentImageResponse

// StartReq starts one agent runtime container when runtime support is available.
type StartReq struct {
	g.Meta `path:"/agents/{id}/start" method:"post" tags:"AgentBox Agents" summary:"Start AgentBox agent runtime" dc:"Start the runtime container for one authenticated-user-owned coding agent. The current migration slice validates ownership and returns runtime-unavailable until trusted container creation, workspace volumes, and tool configuration are migrated."`
	ID     string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
}

// StartRes returns the started coding agent.
type StartRes = AgentInfo

// StopReq stops one agent runtime container when runtime support is available.
type StopReq struct {
	g.Meta `path:"/agents/{id}/stop" method:"post" tags:"AgentBox Agents" summary:"Stop AgentBox agent runtime" dc:"Stop the runtime container for one authenticated-user-owned coding agent. The current migration slice validates ownership and returns runtime-unavailable until trusted runtime lifecycle cleanup is migrated."`
	ID     string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
}

// StopRes returns the stopped coding agent.
type StopRes = AgentInfo

// LogsReq reads runtime logs for one agent when runtime support is available.
type LogsReq struct {
	g.Meta `path:"/agents/{id}/logs" method:"get" tags:"AgentBox Agents" summary:"Get AgentBox agent runtime logs" dc:"Read recent runtime logs for one authenticated-user-owned coding agent. The current migration slice validates ownership and returns runtime-unavailable until trusted container log streaming is migrated."`
	ID     string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
}

// LogsRes returns runtime log text.
type LogsRes = AgentLogsResponse

// DeleteReq deletes one coding agent.
type DeleteReq struct {
	g.Meta        `path:"/agents/{id}" method:"delete" tags:"AgentBox Agents" summary:"Delete AgentBox agent" dc:"Soft-delete one coding agent owned by the authenticated AgentBox user."`
	ID            string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	DeleteVolumes bool   `json:"deleteVolumes" dc:"Whether to remove this agent's managed runtime volumes; currently recorded for later runtime cleanup" eg:"false"`
}

// DeleteRes reports agent deletion state.
type DeleteRes struct {
	Deleted bool `json:"deleted" dc:"Whether the agent was deleted" eg:"true"`
}

// AgentInfo is the public AgentBox coding-agent projection.
type AgentInfo struct {
	ID             string `json:"id" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	UserID         string `json:"userId,omitempty" dc:"AgentBox user ID that owns this agent; omitted when not needed by the client" eg:"usr-admin"`
	Name           string `json:"name" dc:"Agent display name" eg:"Frontend Workbench"`
	ProviderID     int64  `json:"providerId" dc:"Provider ID used by this agent" eg:"1"`
	ProviderName   string `json:"providerName" dc:"Provider display name used by this agent" eg:"OpenAI"`
	ModelName      string `json:"modelName" dc:"Provider model name" eg:"gpt-5"`
	ModelProtocol  string `json:"modelProtocol" dc:"Model protocol: openai=OpenAI-compatible, anthropic=Anthropic-compatible" eg:"openai"`
	ImageID        int64  `json:"imageId" dc:"Coding image ID used by this agent" eg:"1"`
	ImageName      string `json:"imageName" dc:"Coding image display name" eg:"Codex Ubuntu"`
	ImageRef       string `json:"imageRef" dc:"Docker image reference used by this agent" eg:"ghcr.io/example/codex:latest"`
	AgentType      string `json:"agentType" dc:"Abstract agent type: claude_code, codex, custom" eg:"codex"`
	IconKey        string `json:"iconKey" dc:"Optional preset sidebar icon key" eg:"test-tube"`
	Notes          string `json:"notes" dc:"Operator notes for this agent" eg:"Primary frontend coding agent"`
	RuntimeStatus  string `json:"runtimeStatus" dc:"Runtime container status: missing=not found, running=running, stopped=stopped, unavailable=runtime unavailable" eg:"running"`
	ActivityStatus string `json:"activityStatus" dc:"Current activity status projected for workbench controls" eg:"running"`
	ContainerID    string `json:"containerId,omitempty" dc:"AgentBox runtime container ID; omitted when no runtime exists" eg:"ctr-1234567890abcdef"`
	DockerID       string `json:"dockerId,omitempty" dc:"Docker container ID; omitted when no runtime exists" eg:"abcdef1234567890"`
	DeletedAt      *int64 `json:"deletedAt,omitempty" dc:"Soft deletion time as Unix timestamp in milliseconds; omitted when the agent is active" eg:"1704067200000"`
	CreatedAt      int64  `json:"createdAt" dc:"Creation time as Unix timestamp in milliseconds" eg:"1704067200000"`
	UpdatedAt      int64  `json:"updatedAt" dc:"Last update time as Unix timestamp in milliseconds" eg:"1704067201000"`
}

// ChangeAgentImageResponse reports the image switch result.
type ChangeAgentImageResponse struct {
	Agent          AgentInfo `json:"agent" dc:"Updated agent after image switch" eg:"{}"`
	LostPaths      []string  `json:"lostPaths" dc:"Container paths whose image-related volumes may be reset by the image switch" eg:"[\"/etc\",\"/opt\",\"/usr\",\"/var\"]"`
	PreservedPaths []string  `json:"preservedPaths" dc:"Container paths and managed directories preserved by the image switch" eg:"[\"/home\",\"/root\",\"/home/agent/workspace\",\"/home/agent/shared\"]"`
}

// AgentLogsResponse returns Agent runtime logs.
type AgentLogsResponse struct {
	Logs string `json:"logs" dc:"Runtime log text" eg:""`
}
