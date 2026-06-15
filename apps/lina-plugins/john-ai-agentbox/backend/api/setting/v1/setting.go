// This file defines version-one user-scoped setting DTOs for the AgentBox
// plugin. Paths are plugin-relative and are published under
// /x/john-ai-agentbox/api/v1 by source-plugin route registration.

package v1

import "github.com/gogf/gf/v2/frame/g"

// DetailReq gets one current-user persisted setting by key.
type DetailReq struct {
	g.Meta `path:"/settings/{key}" method:"get" tags:"AgentBox Settings" summary:"Get AgentBox setting" dc:"Get one persisted AgentBox setting scoped to the authenticated AgentBox user; missing keys return a structured not-found error so clients can initialize defaults without sharing preferences across users."`
	Key    string `json:"key" v:"required" dc:"Current-user setting key, for example workbench display settings" eg:"workbench"`
}

// DetailRes returns one current-user persisted setting.
type DetailRes = SettingInfo

// UpdateReq upserts one current-user persisted setting by key.
type UpdateReq struct {
	g.Meta `path:"/settings/{key}" method:"put" tags:"AgentBox Settings" summary:"Update AgentBox setting" dc:"Create or update one persisted AgentBox setting value scoped to the authenticated AgentBox user."`
	Key    string `json:"key" v:"required" dc:"Current-user setting key, for example workbench display settings" eg:"workbench"`
	Value  string `json:"value" dc:"Setting value serialized by the caller, usually JSON text for structured settings" eg:"{\"editorFontSize\":14}"`
}

// UpdateRes returns the current-user upserted setting.
type UpdateRes = SettingInfo

// SettingInfo describes one persisted AgentBox setting row scoped to a user.
type SettingInfo struct {
	Key       string `json:"key" dc:"Current-user setting key stored in the key/value settings table" eg:"workbench"`
	Value     string `json:"value" dc:"Setting value serialized by the caller, usually JSON text for structured settings" eg:"{\"editorFontSize\":14}"`
	CreatedAt int64  `json:"createdAt" dc:"Creation time as Unix timestamp in milliseconds" eg:"1704067200000"`
	UpdatedAt int64  `json:"updatedAt" dc:"Last update time as Unix timestamp in milliseconds" eg:"1704067201000"`
}
