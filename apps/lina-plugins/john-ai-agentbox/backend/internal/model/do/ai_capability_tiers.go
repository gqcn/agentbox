// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// AiCapabilityTiers is the golang structure of table john_ai_agentbox_ai_capability_tiers for DAO operations like Where/Data.
type AiCapabilityTiers struct {
	g.Meta      `orm:"table:john_ai_agentbox_ai_capability_tiers, do:true"`
	Code        any        // Capability tier code
	DisplayName any        // Capability tier display name
	Description any        // Capability tier description
	Enabled     any        // Whether the tier is enabled
	CreatedAt   *time.Time // Creation time
	UpdatedAt   *time.Time // Update time
}
