// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// AiCapabilityBindings is the golang structure for table ai_capability_bindings.
type AiCapabilityBindings struct {
	Id              int64      `json:"id"              orm:"id"                description:"Capability binding ID"`
	TierCode        string     `json:"tierCode"        orm:"tier_code"         description:"Capability tier code"`
	ProviderId      int64      `json:"providerId"      orm:"provider_id"       description:"Provider ID"`
	ProviderModelId int64      `json:"providerModelId" orm:"provider_model_id" description:"Provider model ID"`
	Priority        int        `json:"priority"        orm:"priority"          description:"Binding priority"`
	Enabled         bool       `json:"enabled"         orm:"enabled"           description:"Whether the binding is enabled"`
	CreatedAt       *time.Time `json:"createdAt"       orm:"created_at"        description:"Creation time"`
	UpdatedAt       *time.Time `json:"updatedAt"       orm:"updated_at"        description:"Update time"`
}
