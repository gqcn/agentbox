// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// AgentChatSessions is the golang structure of table john_ai_agentbox_agent_chat_sessions for DAO operations like Where/Data.
type AgentChatSessions struct {
	g.Meta             `orm:"table:john_ai_agentbox_agent_chat_sessions, do:true"`
	Id                 any        // Chat session ID
	AgentId            any        // Agent ID
	Title              any        // Session title
	Status             any        // Session status
	ToolType           any        // Connected tool type
	ToolSessionId      any        // Connected tool session ID
	RuntimeState       any        // Runtime state
	LastError          any        // Last runtime error
	MessageCount       any        // Message count
	LastMessagePreview any        // Latest message preview
	CreatedAt          *time.Time // Creation time
	UpdatedAt          *time.Time // Update time
	LastActiveAt       *time.Time // Last activity time
}
