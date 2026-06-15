// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// AgentTerminalSessions is the golang structure for table agent_terminal_sessions.
type AgentTerminalSessions struct {
	Id                 string     `json:"id"                 orm:"id"                   description:"Terminal session ID"`
	UserId             string     `json:"userId"             orm:"user_id"              description:"Owner user ID"`
	AgentId            string     `json:"agentId"            orm:"agent_id"             description:"Agent ID"`
	TerminalId         string     `json:"terminalId"         orm:"terminal_id"          description:"Frontend terminal ID"`
	BackendType        string     `json:"backendType"        orm:"backend_type"         description:"Terminal backend type"`
	BackendSessionName string     `json:"backendSessionName" orm:"backend_session_name" description:"Terminal backend session name"`
	WorkingDir         string     `json:"workingDir"         orm:"working_dir"          description:"Working directory"`
	Shell              string     `json:"shell"              orm:"shell"                description:"Shell path"`
	Status             string     `json:"status"             orm:"status"               description:"Terminal session status"`
	LastError          string     `json:"lastError"          orm:"last_error"           description:"Last terminal error"`
	ClosedAt           *time.Time `json:"closedAt"           orm:"closed_at"            description:"Close time"`
	CreatedAt          *time.Time `json:"createdAt"          orm:"created_at"           description:"Creation time"`
	UpdatedAt          *time.Time `json:"updatedAt"          orm:"updated_at"           description:"Update time"`
}
