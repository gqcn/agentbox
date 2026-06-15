// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// AgentChatMessages is the golang structure of table john_ai_agentbox_agent_chat_messages for DAO operations like Where/Data.
type AgentChatMessages struct {
	g.Meta    `orm:"table:john_ai_agentbox_agent_chat_messages, do:true"`
	Id        any        // Chat message ID
	SessionId any        // Chat session ID
	Sequence  any        // Message sequence in the session
	Role      any        // Message role
	Content   any        // Message content
	Status    any        // Message status
	Metadata  any        // Message metadata JSON
	CreatedAt *time.Time // Creation time
	UpdatedAt *time.Time // Update time
}
