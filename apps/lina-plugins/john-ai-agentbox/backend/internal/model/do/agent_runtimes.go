// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// AgentRuntimes is the golang structure of table john_ai_agentbox_agent_runtimes for DAO operations like Where/Data.
type AgentRuntimes struct {
	g.Meta          `orm:"table:john_ai_agentbox_agent_runtimes, do:true"`
	AgentId         any        // Agent ID
	ContainerId     any        // AgentBox logical container ID
	DockerId        any        // Docker container ID
	Status          any        // Runtime status
	ConfigMountPath any        // Runtime configuration mount path
	CreatedAt       *time.Time // Creation time
	UpdatedAt       *time.Time // Update time
}
