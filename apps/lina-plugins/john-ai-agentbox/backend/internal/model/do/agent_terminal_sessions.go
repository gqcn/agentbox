// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// AgentTerminalSessions is the golang structure of table john_ai_agentbox_agent_terminal_sessions for DAO operations like Where/Data.
type AgentTerminalSessions struct {
	g.Meta             `orm:"table:john_ai_agentbox_agent_terminal_sessions, do:true"`
	Id                 any        // Terminal session ID
	UserId             any        // Owner user ID
	AgentId            any        // Agent ID
	TerminalId         any        // Frontend terminal ID
	BackendType        any        // Terminal backend type
	BackendSessionName any        // Terminal backend session name
	WorkingDir         any        // Working directory
	Shell              any        // Shell path
	Status             any        // Terminal session status
	LastError          any        // Last terminal error
	ClosedAt           *time.Time // Close time
	CreatedAt          *time.Time // Creation time
	UpdatedAt          *time.Time // Update time
}
