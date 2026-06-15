// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// ProviderModels is the golang structure of table john_ai_agentbox_provider_models for DAO operations like Where/Data.
type ProviderModels struct {
	g.Meta       `orm:"table:john_ai_agentbox_provider_models, do:true"`
	Id           any        // Provider model ID
	ProviderId   any        // Provider ID
	Name         any        // Model name
	Protocol     any        // Model protocol
	Source       any        // Model source
	LastSyncedAt *time.Time // Last remote synchronization time
	CreatedAt    *time.Time // Creation time
	UpdatedAt    *time.Time // Update time
}
