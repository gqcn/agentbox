// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// ProviderModels is the golang structure for table provider_models.
type ProviderModels struct {
	Id           int64      `json:"id"           orm:"id"             description:"Provider model ID"`
	ProviderId   int64      `json:"providerId"   orm:"provider_id"    description:"Provider ID"`
	Name         string     `json:"name"         orm:"name"           description:"Model name"`
	Protocol     string     `json:"protocol"     orm:"protocol"       description:"Model protocol"`
	Source       string     `json:"source"       orm:"source"         description:"Model source"`
	LastSyncedAt *time.Time `json:"lastSyncedAt" orm:"last_synced_at" description:"Last remote synchronization time"`
	CreatedAt    *time.Time `json:"createdAt"    orm:"created_at"     description:"Creation time"`
	UpdatedAt    *time.Time `json:"updatedAt"    orm:"updated_at"     description:"Update time"`
}
