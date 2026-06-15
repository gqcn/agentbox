// This file loads AgentBox plugin-owned business configuration during source
// plugin route registration. It uses the existing plugin-scoped config service
// published by the host, and falls back to conservative defaults when the host
// service is unavailable in focused unit tests.

package routes

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	aisvc "john-ai-agentbox/backend/internal/service/ai"
	authsvc "john-ai-agentbox/backend/internal/service/auth"
	catalogsvc "john-ai-agentbox/backend/internal/service/catalog"
	containersvc "john-ai-agentbox/backend/internal/service/container"
	serviceproxysvc "john-ai-agentbox/backend/internal/service/serviceproxy"
	workspacesvc "john-ai-agentbox/backend/internal/service/workspace"

	"lina-core/pkg/plugin/pluginhost"
)

const runtimeModeSingleNode = "single-node"

type pluginConfigReader interface {
	String(ctx context.Context, key string, defaultValue string) (string, error)
	Int(ctx context.Context, key string, defaultValue int) (int, error)
	Duration(ctx context.Context, key string, defaultValue time.Duration) (time.Duration, error)
}

type agentBoxConfig struct {
	RuntimeMode string
	Auth        authsvc.Config
	Catalog     catalogsvc.Config
	AI          aisvc.Config
	Docker      containersvc.RuntimeConfig
	Workspace   workspacesvc.Config
	Service     serviceproxysvc.Config
}

func loadAgentBoxConfig(ctx context.Context, registrar pluginhost.HTTPRegistrar) (agentBoxConfig, error) {
	reader := pluginConfigReaderFromRegistrar(registrar)
	return loadAgentBoxConfigFromReader(ctx, reader)
}

func loadAgentBoxConfigFromReader(ctx context.Context, reader pluginConfigReader) (agentBoxConfig, error) {
	cfg := defaultAgentBoxConfig()
	if reader == nil {
		return cfg, nil
	}
	var err error
	if cfg.RuntimeMode, err = reader.String(ctx, "runtime.mode", cfg.RuntimeMode); err != nil {
		return cfg, err
	}
	cfg.RuntimeMode = strings.TrimSpace(cfg.RuntimeMode)
	if cfg.RuntimeMode == "" {
		cfg.RuntimeMode = runtimeModeSingleNode
	}
	if cfg.RuntimeMode != runtimeModeSingleNode {
		return cfg, gerror.Newf("agentbox runtime.mode is unsupported: %s", cfg.RuntimeMode)
	}

	if cfg.Auth.SessionTTL, err = reader.Duration(ctx, "auth.sessionTtl", cfg.Auth.SessionTTL); err != nil {
		return cfg, err
	}
	if cfg.Catalog.RemoteRequestTimeout, err = reader.Duration(ctx, "providers.requestTimeout", cfg.Catalog.RemoteRequestTimeout); err != nil {
		return cfg, err
	}
	if cfg.Catalog.RemoteModelSyncLimit, err = reader.Int(ctx, "providers.remoteModelSyncLimit", cfg.Catalog.RemoteModelSyncLimit); err != nil {
		return cfg, err
	}
	if cfg.AI.RequestTimeout, err = reader.Duration(ctx, "ai.requestTimeout", cfg.AI.RequestTimeout); err != nil {
		return cfg, err
	}
	if cfg.Docker.Host, err = reader.String(ctx, "runtime.docker.host", cfg.Docker.Host); err != nil {
		return cfg, err
	}
	if cfg.Docker.ContainerLogTail, err = reader.Int(ctx, "runtime.docker.containerLogTail", cfg.Docker.ContainerLogTail); err != nil {
		return cfg, err
	}
	if cfg.Docker.StopTimeout, err = reader.Duration(ctx, "runtime.docker.stopTimeout", cfg.Docker.StopTimeout); err != nil {
		return cfg, err
	}
	if cfg.Workspace.WorkspaceRootPath, err = reader.String(ctx, "runtime.workspace.rootPath", cfg.Workspace.WorkspaceRootPath); err != nil {
		return cfg, err
	}
	if cfg.Workspace.SharedRootPath, err = reader.String(ctx, "runtime.workspace.sharedPath", cfg.Workspace.SharedRootPath); err != nil {
		return cfg, err
	}
	if cfg.Workspace.PreviewLimitBytes, err = readInt64Config(ctx, reader, "runtime.workspace.previewLimitBytes", cfg.Workspace.PreviewLimitBytes); err != nil {
		return cfg, err
	}
	if cfg.Workspace.UploadFileLimitBytes, err = readInt64Config(ctx, reader, "runtime.workspace.uploadFileLimitBytes", cfg.Workspace.UploadFileLimitBytes); err != nil {
		return cfg, err
	}
	if cfg.Workspace.UploadCountLimit, err = reader.Int(ctx, "runtime.workspace.uploadCountLimit", cfg.Workspace.UploadCountLimit); err != nil {
		return cfg, err
	}
	if cfg.Workspace.SkillListLimit, err = reader.Int(ctx, "runtime.workspace.skillListLimit", cfg.Workspace.SkillListLimit); err != nil {
		return cfg, err
	}
	if cfg.Workspace.SkillManifestLimitBytes, err = readInt64Config(ctx, reader, "runtime.workspace.skillManifestLimitBytes", cfg.Workspace.SkillManifestLimitBytes); err != nil {
		return cfg, err
	}
	if cfg.Service.RuntimeServiceListLimit, err = reader.Int(ctx, "runtime.services.discoveryLimit", cfg.Service.RuntimeServiceListLimit); err != nil {
		return cfg, err
	}
	cfg.Docker.Workspace = cfg.Workspace
	cfg.Docker.Service = cfg.Service
	return cfg, nil
}

func defaultAgentBoxConfig() agentBoxConfig {
	workspaceConfig := workspacesvc.DefaultConfig()
	serviceConfig := serviceproxysvc.DefaultConfig()
	return agentBoxConfig{
		RuntimeMode: runtimeModeSingleNode,
		Auth:        authsvc.Config{SessionTTL: 24 * time.Hour},
		Catalog: catalogsvc.Config{
			RemoteModelSyncLimit: 200,
			RemoteRequestTimeout: 20 * time.Second,
		},
		AI: aisvc.Config{RequestTimeout: 20 * time.Second},
		Docker: containersvc.RuntimeConfig{
			ContainerLogTail: 400,
			StopTimeout:      10 * time.Second,
			Workspace:        workspaceConfig,
			Service:          serviceConfig,
		},
		Workspace: workspaceConfig,
		Service:   serviceConfig,
	}
}

func pluginConfigReaderFromRegistrar(registrar pluginhost.HTTPRegistrar) pluginConfigReader {
	if registrar == nil || registrar.Services() == nil || registrar.Services().Plugins() == nil {
		return nil
	}
	return registrar.Services().Plugins().Config()
}

func readInt64Config(ctx context.Context, reader pluginConfigReader, key string, defaultValue int64) (int64, error) {
	value, err := reader.Int(ctx, key, int(defaultValue))
	if err != nil {
		return 0, err
	}
	return int64(value), nil
}
