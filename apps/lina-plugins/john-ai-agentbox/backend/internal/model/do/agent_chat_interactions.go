// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// AgentChatInteractions is the golang structure of table john_ai_agentbox_agent_chat_interactions for DAO operations like Where/Data.
type AgentChatInteractions struct {
	g.Meta             `orm:"table:john_ai_agentbox_agent_chat_interactions, do:true"`
	Id                 any        // Interaction ID
	AgentId            any        // Agent ID
	SessionId          any        // Chat session ID
	AssistantMessageId any        // Related assistant message ID
	ToolType           any        // Tool type
	ToolInteractionId  any        // External tool interaction ID
	InteractionType    any        // Interaction type
	Status             any        // Interaction status
	Title              any        // Interaction title
	Body               any        // Interaction body
	RiskLevel          any        // Risk level
	PayloadJson        any        // Interaction payload JSON
	ResponseJson       any        // Interaction response JSON
	ResponseMode       any        // Response mode
	ResponseScope      any        // Response scope
	ExpiresAt          *time.Time // Expiration time
	ResolvedAt         *time.Time // Resolution time
	CreatedAt          *time.Time // Creation time
	UpdatedAt          *time.Time // Update time
	DeletedAt          *time.Time // Soft deletion time
}
