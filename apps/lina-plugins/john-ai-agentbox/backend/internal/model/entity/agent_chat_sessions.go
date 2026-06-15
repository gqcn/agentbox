// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// AgentChatSessions is the golang structure for table agent_chat_sessions.
type AgentChatSessions struct {
	Id                 string     `json:"id"                 orm:"id"                   description:"Chat session ID"`
	AgentId            string     `json:"agentId"            orm:"agent_id"             description:"Agent ID"`
	Title              string     `json:"title"              orm:"title"                description:"Session title"`
	Status             string     `json:"status"             orm:"status"               description:"Session status"`
	ToolType           string     `json:"toolType"           orm:"tool_type"            description:"Connected tool type"`
	ToolSessionId      string     `json:"toolSessionId"      orm:"tool_session_id"      description:"Connected tool session ID"`
	RuntimeState       string     `json:"runtimeState"       orm:"runtime_state"        description:"Runtime state"`
	LastError          string     `json:"lastError"          orm:"last_error"           description:"Last runtime error"`
	MessageCount       int64      `json:"messageCount"       orm:"message_count"        description:"Message count"`
	LastMessagePreview string     `json:"lastMessagePreview" orm:"last_message_preview" description:"Latest message preview"`
	CreatedAt          *time.Time `json:"createdAt"          orm:"created_at"           description:"Creation time"`
	UpdatedAt          *time.Time `json:"updatedAt"          orm:"updated_at"           description:"Update time"`
	LastActiveAt       *time.Time `json:"lastActiveAt"       orm:"last_active_at"       description:"Last activity time"`
}
