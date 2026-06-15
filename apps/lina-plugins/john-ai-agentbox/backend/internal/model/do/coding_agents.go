// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// CodingAgents is the golang structure of table john_ai_agentbox_coding_agents for DAO operations like Where/Data.
type CodingAgents struct {
	g.Meta        `orm:"table:john_ai_agentbox_coding_agents, do:true"`
	Id            any        // Agent ID
	UserId        any        // Owner user ID
	Name          any        // Agent display name
	ProviderId    any        // Provider ID
	ModelName     any        // Selected model name
	ModelProtocol any        // Selected model protocol
	ImageId       any        // Coding image ID
	AgentType     any        // Agent runtime type
	IconKey       any        // Agent icon key
	Notes         any        // Agent notes
	DeletedAt     *time.Time // Soft deletion time
	CreatedAt     *time.Time // Creation time
	UpdatedAt     *time.Time // Update time
}
