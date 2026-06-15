// This file defines version-one container and Docker health DTOs for the
// AgentBox plugin. Paths are plugin-relative and are published under
// /x/john-ai-agentbox/api/v1 by source-plugin route registration.

package v1

import "github.com/gogf/gf/v2/frame/g"

// DockerHealthReq checks whether the AgentBox runtime backend is available.
type DockerHealthReq struct {
	g.Meta `path:"/health/docker" method:"get" tags:"AgentBox Containers" summary:"Check AgentBox Docker runtime health" dc:"Check whether the Docker runtime backend is reachable for the authenticated AgentBox user. Container lifecycle operations are separately scoped by AgentBox plugin labels and the current user label."`
}

// DockerHealthRes returns Docker runtime health details.
type DockerHealthRes = DockerHealthResponse

// ListReq lists managed runtime containers for the current AgentBox user.
type ListReq struct {
	g.Meta `path:"/containers" method:"get" tags:"AgentBox Containers" summary:"List AgentBox runtime containers" dc:"List Docker runtime containers that carry AgentBox plugin labels and belong to the authenticated AgentBox user. Unlabelled containers and containers owned by other AgentBox users are never returned."`
}

// ListRes returns managed runtime containers.
type ListRes = []ContainerInfo

// CreateReq creates one managed runtime container for the current AgentBox user.
type CreateReq struct {
	g.Meta `path:"/containers" method:"post" tags:"AgentBox Containers" summary:"Create AgentBox runtime container" dc:"Create one AgentBox managed runtime container for the authenticated AgentBox user. Creation remains unavailable until trusted runtime image and configuration migration is complete."`
	Name   string `json:"name" dc:"Display name for the runtime container" eg:"frontend-workbench"`
}

// CreateRes returns the created runtime container.
type CreateRes = ContainerInfo

// DetailReq gets one managed runtime container for the current AgentBox user.
type DetailReq struct {
	g.Meta `path:"/containers/{id}" method:"get" tags:"AgentBox Containers" summary:"Get AgentBox runtime container" dc:"Get one Docker runtime container only when it carries AgentBox plugin labels and belongs to the authenticated AgentBox user. Missing containers and containers owned by other users return the same not found semantics."`
	ID     string `json:"id" v:"required" dc:"AgentBox container ID or runtime container ID" eg:"ctr-abc123def456"`
}

// DetailRes returns one managed runtime container.
type DetailRes = ContainerInfo

// StartReq starts one managed runtime container for the current AgentBox user.
type StartReq struct {
	g.Meta `path:"/containers/{id}/start" method:"post" tags:"AgentBox Containers" summary:"Start AgentBox runtime container" dc:"Start one Docker runtime container only when it carries AgentBox plugin labels and belongs to the authenticated AgentBox user. Missing containers and containers owned by other users return the same not found semantics."`
	ID     string `json:"id" v:"required" dc:"AgentBox container ID or runtime container ID" eg:"ctr-abc123def456"`
}

// StartRes returns the started runtime container.
type StartRes = ContainerInfo

// StopReq stops one managed runtime container for the current AgentBox user.
type StopReq struct {
	g.Meta `path:"/containers/{id}/stop" method:"post" tags:"AgentBox Containers" summary:"Stop AgentBox runtime container" dc:"Stop one Docker runtime container only when it carries AgentBox plugin labels and belongs to the authenticated AgentBox user. Missing containers and containers owned by other users return the same not found semantics."`
	ID     string `json:"id" v:"required" dc:"AgentBox container ID or runtime container ID" eg:"ctr-abc123def456"`
}

// StopRes returns the stopped runtime container.
type StopRes = ContainerInfo

// DeleteReq deletes one managed runtime container for the current AgentBox user.
type DeleteReq struct {
	g.Meta `path:"/containers/{id}" method:"delete" tags:"AgentBox Containers" summary:"Delete AgentBox runtime container" dc:"Delete one Docker runtime container only when it carries AgentBox plugin labels and belongs to the authenticated AgentBox user. Missing containers and containers owned by other users return the same not found semantics."`
	ID     string `json:"id" v:"required" dc:"AgentBox container ID or runtime container ID" eg:"ctr-abc123def456"`
}

// DeleteRes reports deletion state.
type DeleteRes struct {
	Deleted bool `json:"deleted" dc:"Whether the container was deleted" eg:"true"`
}

// LogsReq reads logs from one managed runtime container for the current AgentBox user.
type LogsReq struct {
	g.Meta `path:"/containers/{id}/logs" method:"get" tags:"AgentBox Containers" summary:"Get AgentBox runtime container logs" dc:"Read recent logs for one Docker runtime container only when it carries AgentBox plugin labels and belongs to the authenticated AgentBox user. Missing containers and containers owned by other users return the same not found semantics."`
	ID     string `json:"id" v:"required" dc:"AgentBox container ID or runtime container ID" eg:"ctr-abc123def456"`
}

// LogsRes returns runtime container logs.
type LogsRes = LogsResponse

// DockerHealthResponse describes Docker runtime health.
type DockerHealthResponse struct {
	OK         bool   `json:"ok" dc:"Whether the runtime backend is reachable" eg:"false"`
	APIVersion string `json:"apiVersion,omitempty" dc:"Docker API version when available" eg:"1.45"`
	OSType     string `json:"osType,omitempty" dc:"Docker daemon operating system type when available" eg:"linux"`
	Error      string `json:"error,omitempty" dc:"Runtime health error message" eg:"runtime unavailable"`
}

// MountInfo describes one runtime container mount.
type MountInfo struct {
	Type        string `json:"type" dc:"Mount type" eg:"volume"`
	Source      string `json:"source" dc:"Mount source" eg:"agentbox-home"`
	Destination string `json:"destination" dc:"Mount destination" eg:"/home/agent/shared"`
	Name        string `json:"name,omitempty" dc:"Optional mount name" eg:"agentbox-home"`
}

// ContainerInfo describes one AgentBox managed runtime container.
type ContainerInfo struct {
	ID        string            `json:"id" dc:"AgentBox container ID" eg:"ctr-abc123def456"`
	Name      string            `json:"name" dc:"Container display name" eg:"frontend-workbench"`
	DockerID  string            `json:"dockerId" dc:"Runtime container ID" eg:"8b1a9953c461"`
	Image     string            `json:"image" dc:"Container image reference" eg:"ghcr.io/example/agentbox:latest"`
	State     string            `json:"state" dc:"Container state" eg:"running"`
	Status    string            `json:"status" dc:"Container status text" eg:"Up 5 minutes"`
	CreatedAt int64             `json:"createdAt" dc:"Container creation time as Unix timestamp in milliseconds" eg:"1704067200000"`
	Mounts    []MountInfo       `json:"mounts,omitempty" dc:"Container mounts" eg:"[]"`
	Labels    map[string]string `json:"labels,omitempty" dc:"Container labels" eg:"{}"`
	Workspace string            `json:"workspace,omitempty" dc:"Container workspace mount path" eg:"/home/agent/workspace"`
}

// LogsResponse returns runtime logs.
type LogsResponse struct {
	Logs string `json:"logs" dc:"Runtime log text" eg:""`
}
