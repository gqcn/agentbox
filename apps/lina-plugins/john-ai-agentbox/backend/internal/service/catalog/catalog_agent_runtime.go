// This file owns the migrated Agent runtime lifecycle boundary. The current
// slice creates and controls one plugin-labelled long-lived Docker container
// per Agent, then persists the trusted Agent-to-container mapping.

package catalog

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"

	"john-ai-agentbox/backend/internal/dao"
	"john-ai-agentbox/backend/internal/model/do"
	"john-ai-agentbox/backend/internal/model/entity"
)

const (
	// AgentRuntimeStatusRunning marks a runtime container that should accept
	// later shell, workspace, and service-proxy operations.
	AgentRuntimeStatusRunning = "running"
	// AgentRuntimeStatusStopped marks an existing runtime container that is not running.
	AgentRuntimeStatusStopped = "stopped"
)

// StartUserAgentRuntime validates ownership, creates or starts the plugin-managed Agent runtime container, and persists the mapping.
func (s *serviceImpl) StartUserAgentRuntime(ctx context.Context, userID string, agentID string) (*AgentInfo, error) {
	agent, err := s.GetUserAgent(ctx, userID, agentID)
	if err != nil {
		return nil, err
	}
	if s.runtimeBackend == nil {
		return nil, bizerr.NewCode(CodeCatalogRuntimeUnavailable)
	}
	runtime, err := s.getAgentRuntime(ctx, agent.ID)
	if err != nil {
		return nil, err
	}
	var info *AgentRuntimeContainerInfo
	if runtime == nil || strings.TrimSpace(runtime.ContainerId) == "" {
		info, err = s.runtimeBackend.CreateAgentRuntime(ctx, AgentRuntimeContainerInput{
			AgentID:   agent.ID,
			UserID:    agent.UserID,
			Name:      agent.Name,
			ImageID:   agent.ImageID,
			ImageRef:  agent.ImageRef,
			AgentType: agent.AgentType,
		})
	} else {
		info, err = s.runtimeBackend.StartAgentRuntime(ctx, agent.UserID, agent.ID, runtime.ContainerId)
	}
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeCatalogRuntimeUnavailable)
	}
	if err := s.upsertAgentRuntime(ctx, agent.ID, info); err != nil {
		return nil, err
	}
	return s.GetUserAgent(ctx, userID, agentID)
}

// StopUserAgentRuntime validates ownership, stops the plugin-managed Agent runtime container, and persists stopped status.
func (s *serviceImpl) StopUserAgentRuntime(ctx context.Context, userID string, agentID string) (*AgentInfo, error) {
	agent, err := s.GetUserAgent(ctx, userID, agentID)
	if err != nil {
		return nil, err
	}
	if s.runtimeBackend == nil {
		return nil, bizerr.NewCode(CodeCatalogRuntimeUnavailable)
	}
	runtime, err := s.getAgentRuntime(ctx, agent.ID)
	if err != nil {
		return nil, err
	}
	if runtime == nil || strings.TrimSpace(runtime.ContainerId) == "" {
		return nil, bizerr.NewCode(CodeCatalogRuntimeUnavailable)
	}
	info, err := s.runtimeBackend.StopAgentRuntime(ctx, agent.UserID, agent.ID, runtime.ContainerId)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeCatalogRuntimeUnavailable)
	}
	if err := s.upsertAgentRuntime(ctx, agent.ID, info); err != nil {
		return nil, err
	}
	return s.GetUserAgent(ctx, userID, agentID)
}

// UserAgentRuntimeLogs validates ownership and reads recent logs from the plugin-managed Agent runtime container.
func (s *serviceImpl) UserAgentRuntimeLogs(ctx context.Context, userID string, agentID string) (*AgentLogsOutput, error) {
	agent, err := s.GetUserAgent(ctx, userID, agentID)
	if err != nil {
		return nil, err
	}
	if s.runtimeBackend == nil {
		return nil, bizerr.NewCode(CodeCatalogRuntimeUnavailable)
	}
	runtime, err := s.getAgentRuntime(ctx, agent.ID)
	if err != nil {
		return nil, err
	}
	if runtime == nil || strings.TrimSpace(runtime.ContainerId) == "" {
		return nil, bizerr.NewCode(CodeCatalogRuntimeUnavailable)
	}
	output, err := s.runtimeBackend.AgentRuntimeLogs(ctx, agent.UserID, agent.ID, runtime.ContainerId)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeCatalogRuntimeUnavailable)
	}
	if output == nil {
		return &AgentLogsOutput{}, nil
	}
	return &AgentLogsOutput{Logs: output.Logs}, nil
}

func (s *serviceImpl) getAgentRuntime(ctx context.Context, agentID string) (*entity.AgentRuntimes, error) {
	var row *entity.AgentRuntimes
	err := dao.AgentRuntimes.Ctx(ctx).
		Where(do.AgentRuntimes{AgentId: strings.TrimSpace(agentID)}).
		Scan(&row)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeCatalogStoreUnavailable)
	}
	return row, nil
}

func (s *serviceImpl) upsertAgentRuntime(ctx context.Context, agentID string, info *AgentRuntimeContainerInfo) error {
	agentID = strings.TrimSpace(agentID)
	if agentID == "" || info == nil || strings.TrimSpace(info.ContainerID) == "" || strings.TrimSpace(info.DockerID) == "" {
		return bizerr.NewCode(CodeCatalogRuntimeUnavailable)
	}
	status := normalizeAgentRuntimeStatus(info.Status)
	return dao.AgentRuntimes.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		count, err := dao.AgentRuntimes.Ctx(ctx).
			Where(do.AgentRuntimes{AgentId: agentID}).
			Count()
		if err != nil {
			return bizerr.WrapCode(err, CodeCatalogStoreUnavailable)
		}
		data := do.AgentRuntimes{
			ContainerId: strings.TrimSpace(info.ContainerID),
			DockerId:    strings.TrimSpace(info.DockerID),
			Status:      status,
		}
		if count == 0 {
			data.AgentId = agentID
			if _, err := dao.AgentRuntimes.Ctx(ctx).Data(data).Insert(); err != nil {
				return bizerr.WrapCode(err, CodeCatalogStoreUnavailable)
			}
			return nil
		}
		if _, err := dao.AgentRuntimes.Ctx(ctx).
			Where(do.AgentRuntimes{AgentId: agentID}).
			Data(data).
			Update(); err != nil {
			return bizerr.WrapCode(err, CodeCatalogStoreUnavailable)
		}
		return nil
	})
}

func normalizeAgentRuntimeStatus(status string) string {
	switch strings.TrimSpace(status) {
	case AgentRuntimeStatusRunning:
		return AgentRuntimeStatusRunning
	case AgentRuntimeStatusStopped:
		return AgentRuntimeStatusStopped
	default:
		return AgentRuntimeStatusStopped
	}
}
