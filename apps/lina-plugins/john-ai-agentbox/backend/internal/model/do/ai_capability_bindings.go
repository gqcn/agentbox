// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// AiCapabilityBindings is the golang structure of table john_ai_agentbox_ai_capability_bindings for DAO operations like Where/Data.
type AiCapabilityBindings struct {
	g.Meta          `orm:"table:john_ai_agentbox_ai_capability_bindings, do:true"`
	Id              any        // Capability binding ID
	TierCode        any        // Capability tier code
	ProviderId      any        // Provider ID
	ProviderModelId any        // Provider model ID
	Priority        any        // Binding priority
	Enabled         any        // Whether the binding is enabled
	CreatedAt       *time.Time // Creation time
	UpdatedAt       *time.Time // Update time
}
