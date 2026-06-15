// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// AiCapabilityTiers is the golang structure for table ai_capability_tiers.
type AiCapabilityTiers struct {
	Code        string     `json:"code"        orm:"code"         description:"Capability tier code"`
	DisplayName string     `json:"displayName" orm:"display_name" description:"Capability tier display name"`
	Description string     `json:"description" orm:"description"  description:"Capability tier description"`
	Enabled     bool       `json:"enabled"     orm:"enabled"      description:"Whether the tier is enabled"`
	CreatedAt   *time.Time `json:"createdAt"   orm:"created_at"   description:"Creation time"`
	UpdatedAt   *time.Time `json:"updatedAt"   orm:"updated_at"   description:"Update time"`
}
