// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// CodingAgents is the golang structure for table coding_agents.
type CodingAgents struct {
	Id            string     `json:"id"            orm:"id"             description:"Agent ID"`
	UserId        string     `json:"userId"        orm:"user_id"        description:"Owner user ID"`
	Name          string     `json:"name"          orm:"name"           description:"Agent display name"`
	ProviderId    int64      `json:"providerId"    orm:"provider_id"    description:"Provider ID"`
	ModelName     string     `json:"modelName"     orm:"model_name"     description:"Selected model name"`
	ModelProtocol string     `json:"modelProtocol" orm:"model_protocol" description:"Selected model protocol"`
	ImageId       int64      `json:"imageId"       orm:"image_id"       description:"Coding image ID"`
	AgentType     string     `json:"agentType"     orm:"agent_type"     description:"Agent runtime type"`
	IconKey       string     `json:"iconKey"       orm:"icon_key"       description:"Agent icon key"`
	Notes         string     `json:"notes"         orm:"notes"          description:"Agent notes"`
	DeletedAt     *time.Time `json:"deletedAt"     orm:"deleted_at"     description:"Soft deletion time"`
	CreatedAt     *time.Time `json:"createdAt"     orm:"created_at"     description:"Creation time"`
	UpdatedAt     *time.Time `json:"updatedAt"     orm:"updated_at"     description:"Update time"`
}
