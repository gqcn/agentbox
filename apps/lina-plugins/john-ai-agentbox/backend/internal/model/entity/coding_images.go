// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// CodingImages is the golang structure for table coding_images.
type CodingImages struct {
	Id           int64      `json:"id"           orm:"id"            description:"Coding image ID"`
	Name         string     `json:"name"         orm:"name"          description:"Image display name"`
	ImageRef     string     `json:"imageRef"     orm:"image_ref"     description:"Container image reference"`
	AgentType    string     `json:"agentType"    orm:"agent_type"    description:"Agent runtime type"`
	DefaultShell string     `json:"defaultShell" orm:"default_shell" description:"Default shell path"`
	Notes        string     `json:"notes"        orm:"notes"         description:"Image notes"`
	Enabled      bool       `json:"enabled"      orm:"enabled"       description:"Whether the image is enabled"`
	IsDefault    bool       `json:"isDefault"    orm:"is_default"    description:"Whether the image is a default option"`
	CreatedAt    *time.Time `json:"createdAt"    orm:"created_at"    description:"Creation time"`
	UpdatedAt    *time.Time `json:"updatedAt"    orm:"updated_at"    description:"Update time"`
}
