// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// AgentChatInteractions is the golang structure for table agent_chat_interactions.
type AgentChatInteractions struct {
	Id                 string     `json:"id"                 orm:"id"                   description:"Interaction ID"`
	AgentId            string     `json:"agentId"            orm:"agent_id"             description:"Agent ID"`
	SessionId          string     `json:"sessionId"          orm:"session_id"           description:"Chat session ID"`
	AssistantMessageId int64      `json:"assistantMessageId" orm:"assistant_message_id" description:"Related assistant message ID"`
	ToolType           string     `json:"toolType"           orm:"tool_type"            description:"Tool type"`
	ToolInteractionId  string     `json:"toolInteractionId"  orm:"tool_interaction_id"  description:"External tool interaction ID"`
	InteractionType    string     `json:"interactionType"    orm:"interaction_type"     description:"Interaction type"`
	Status             string     `json:"status"             orm:"status"               description:"Interaction status"`
	Title              string     `json:"title"              orm:"title"                description:"Interaction title"`
	Body               string     `json:"body"               orm:"body"                 description:"Interaction body"`
	RiskLevel          string     `json:"riskLevel"          orm:"risk_level"           description:"Risk level"`
	PayloadJson        string     `json:"payloadJson"        orm:"payload_json"         description:"Interaction payload JSON"`
	ResponseJson       string     `json:"responseJson"       orm:"response_json"        description:"Interaction response JSON"`
	ResponseMode       string     `json:"responseMode"       orm:"response_mode"        description:"Response mode"`
	ResponseScope      string     `json:"responseScope"      orm:"response_scope"       description:"Response scope"`
	ExpiresAt          *time.Time `json:"expiresAt"          orm:"expires_at"           description:"Expiration time"`
	ResolvedAt         *time.Time `json:"resolvedAt"         orm:"resolved_at"          description:"Resolution time"`
	CreatedAt          *time.Time `json:"createdAt"          orm:"created_at"           description:"Creation time"`
	UpdatedAt          *time.Time `json:"updatedAt"          orm:"updated_at"           description:"Update time"`
	DeletedAt          *time.Time `json:"deletedAt"          orm:"deleted_at"           description:"Soft deletion time"`
}
