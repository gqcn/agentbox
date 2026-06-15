// This file defines version-one coding-image DTOs for the AgentBox plugin.
// Paths are plugin-relative and are published under /x/john-ai-agentbox/api/v1
// by source-plugin route registration.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListReq lists coding images.
type ListReq struct {
	g.Meta `path:"/images" method:"get" tags:"AgentBox Images" summary:"List AgentBox images" dc:"List coding container images available for AgentBox agent creation."`
}

// ListRes returns coding images.
type ListRes = []CodingImageInfo

// CreateReq creates a coding image.
type CreateReq struct {
	g.Meta       `path:"/images" method:"post" tags:"AgentBox Images" summary:"Create AgentBox image" dc:"Create a coding image profile for one abstract AgentBox agent type."`
	Name         string `json:"name" v:"required" dc:"Image display name" eg:"Codex Ubuntu"`
	ImageRef     string `json:"imageRef" v:"required" dc:"Docker image reference" eg:"ghcr.io/example/codex:latest"`
	AgentType    string `json:"agentType" v:"required" dc:"Supported agent type: claude_code, codex, custom" eg:"codex"`
	DefaultShell string `json:"defaultShell" dc:"Default shell used for terminal sessions" eg:"/bin/bash"`
	Notes        string `json:"notes" dc:"Operator notes for this image" eg:"Preloaded Codex CLI image"`
	Enabled      bool   `json:"enabled" dc:"Whether this image can be selected by new agents" eg:"true"`
}

// CreateRes returns the created coding image.
type CreateRes = CodingImageInfo

// UpdateReq updates a coding image.
type UpdateReq struct {
	g.Meta       `path:"/images/{id}" method:"put" tags:"AgentBox Images" summary:"Update AgentBox image" dc:"Update a coding image profile for one abstract AgentBox agent type."`
	ID           int64  `json:"id" v:"required|min:1" dc:"Coding image ID" eg:"1"`
	Name         string `json:"name" v:"required" dc:"Image display name" eg:"Codex Ubuntu"`
	ImageRef     string `json:"imageRef" v:"required" dc:"Docker image reference" eg:"ghcr.io/example/codex:latest"`
	AgentType    string `json:"agentType" v:"required" dc:"Supported agent type: claude_code, codex, custom" eg:"codex"`
	DefaultShell string `json:"defaultShell" dc:"Default shell used for terminal sessions" eg:"/bin/bash"`
	Notes        string `json:"notes" dc:"Operator notes for this image" eg:"Preloaded Codex CLI image"`
	Enabled      bool   `json:"enabled" dc:"Whether this image can be selected by new agents" eg:"true"`
}

// UpdateRes returns the updated coding image.
type UpdateRes = CodingImageInfo

// DeleteReq deletes a coding image.
type DeleteReq struct {
	g.Meta `path:"/images/{id}" method:"delete" tags:"AgentBox Images" summary:"Delete AgentBox image" dc:"Delete a coding image if no AgentBox agent references it and it is not the default image."`
	ID     int64 `json:"id" v:"required|min:1" dc:"Coding image ID" eg:"1"`
}

// DeleteRes reports image deletion state.
type DeleteRes struct {
	Deleted bool `json:"deleted" dc:"Whether the coding image was deleted" eg:"true"`
}

// CodingImageInfo is the public AgentBox coding-image projection.
type CodingImageInfo struct {
	ID           int64  `json:"id" dc:"Coding image ID" eg:"1"`
	Name         string `json:"name" dc:"Image display name" eg:"Codex Ubuntu"`
	ImageRef     string `json:"imageRef" dc:"Docker image reference" eg:"ghcr.io/example/codex:latest"`
	AgentType    string `json:"agentType" dc:"Supported agent type: claude_code, codex, custom" eg:"codex"`
	DefaultShell string `json:"defaultShell" dc:"Default shell used for terminal sessions" eg:"/bin/bash"`
	Notes        string `json:"notes" dc:"Operator notes for this image" eg:"Preloaded Codex CLI image"`
	Enabled      bool   `json:"enabled" dc:"Whether this image can be selected by new agents" eg:"true"`
	IsDefault    bool   `json:"isDefault" dc:"Whether this image is the system default coding image" eg:"false"`
	CreatedAt    int64  `json:"createdAt" dc:"Creation time as Unix timestamp in milliseconds" eg:"1704067200000"`
	UpdatedAt    int64  `json:"updatedAt" dc:"Last update time as Unix timestamp in milliseconds" eg:"1704067201000"`
}
