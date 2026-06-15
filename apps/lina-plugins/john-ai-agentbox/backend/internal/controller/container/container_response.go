// This file maps AgentBox container service projections to public DTOs while
// keeping runtime backend internals out of HTTP response contracts.

package container

import (
	v1 "john-ai-agentbox/backend/api/container/v1"
	containersvc "john-ai-agentbox/backend/internal/service/container"
)

func toDockerHealthResponse(item *containersvc.DockerHealthResponse) *v1.DockerHealthResponse {
	if item == nil {
		return nil
	}
	return &v1.DockerHealthResponse{
		OK:         item.OK,
		APIVersion: item.APIVersion,
		OSType:     item.OSType,
		Error:      item.Error,
	}
}

func toContainerListResponse(items []containersvc.ContainerInfo) []v1.ContainerInfo {
	out := make([]v1.ContainerInfo, 0, len(items))
	for _, item := range items {
		out = append(out, toContainerResponse(item))
	}
	return out
}

func toContainerResponse(item containersvc.ContainerInfo) v1.ContainerInfo {
	return v1.ContainerInfo{
		ID:        item.ID,
		Name:      item.Name,
		DockerID:  item.DockerID,
		Image:     item.Image,
		State:     item.State,
		Status:    item.Status,
		CreatedAt: item.CreatedAt,
		Mounts:    toMountListResponse(item.Mounts),
		Labels:    item.Labels,
		Workspace: item.Workspace,
	}
}

func toMountListResponse(items []containersvc.MountInfo) []v1.MountInfo {
	out := make([]v1.MountInfo, 0, len(items))
	for _, item := range items {
		out = append(out, v1.MountInfo{
			Type:        item.Type,
			Source:      item.Source,
			Destination: item.Destination,
			Name:        item.Name,
		})
	}
	return out
}

func toLogsResponse(item *containersvc.LogsResponse) *v1.LogsResponse {
	if item == nil {
		return nil
	}
	return &v1.LogsResponse{Logs: item.Logs}
}
