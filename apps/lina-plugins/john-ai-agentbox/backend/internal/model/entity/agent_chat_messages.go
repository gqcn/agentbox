// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// AgentChatMessages is the golang structure for table agent_chat_messages.
type AgentChatMessages struct {
	Id        int64      `json:"id"        orm:"id"         description:"Chat message ID"`
	SessionId string     `json:"sessionId" orm:"session_id" description:"Chat session ID"`
	Sequence  int64      `json:"sequence"  orm:"sequence"   description:"Message sequence in the session"`
	Role      string     `json:"role"      orm:"role"       description:"Message role"`
	Content   string     `json:"content"   orm:"content"    description:"Message content"`
	Status    string     `json:"status"    orm:"status"     description:"Message status"`
	Metadata  string     `json:"metadata"  orm:"metadata"   description:"Message metadata JSON"`
	CreatedAt *time.Time `json:"createdAt" orm:"created_at" description:"Creation time"`
	UpdatedAt *time.Time `json:"updatedAt" orm:"updated_at" description:"Update time"`
}
