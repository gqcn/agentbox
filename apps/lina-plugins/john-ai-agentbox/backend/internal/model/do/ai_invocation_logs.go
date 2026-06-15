// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// AiInvocationLogs is the golang structure of table john_ai_agentbox_ai_invocation_logs for DAO operations like Where/Data.
type AiInvocationLogs struct {
	g.Meta          `orm:"table:john_ai_agentbox_ai_invocation_logs, do:true"`
	Id              any        // AI invocation log ID
	Purpose         any        // Invocation purpose
	TierCode        any        // Capability tier code
	ProviderId      any        // Provider ID
	ProviderModelId any        // Provider model ID
	ModelName       any        // Model name used for the invocation
	Protocol        any        // Model protocol
	Status          any        // Invocation status
	LatencyMs       any        // Invocation latency in milliseconds
	ErrorMessage    any        // Invocation error message
	CreatedAt       *time.Time // Creation time
}
