// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// AgentRuntimes is the golang structure for table agent_runtimes.
type AgentRuntimes struct {
	AgentId         string     `json:"agentId"         orm:"agent_id"          description:"Agent ID"`
	ContainerId     string     `json:"containerId"     orm:"container_id"      description:"AgentBox logical container ID"`
	DockerId        string     `json:"dockerId"        orm:"docker_id"         description:"Docker container ID"`
	Status          string     `json:"status"          orm:"status"            description:"Runtime status"`
	ConfigMountPath string     `json:"configMountPath" orm:"config_mount_path" description:"Runtime configuration mount path"`
	CreatedAt       *time.Time `json:"createdAt"       orm:"created_at"        description:"Creation time"`
	UpdatedAt       *time.Time `json:"updatedAt"       orm:"updated_at"        description:"Update time"`
}
