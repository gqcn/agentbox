// Package routes wires AgentBox source-plugin HTTP routes into the Lina host
// registrar. It owns only startup-time composition; business logic remains in
// backend/internal controller and service packages.
package routes

import (
	"context"
	"io/fs"
	"net/http"

	agentcontroller "john-ai-agentbox/backend/internal/controller/agent"
	aicontroller "john-ai-agentbox/backend/internal/controller/ai"
	authcontroller "john-ai-agentbox/backend/internal/controller/auth"
	containercontroller "john-ai-agentbox/backend/internal/controller/container"
	gatewaycontroller "john-ai-agentbox/backend/internal/controller/gateway"
	imagecontroller "john-ai-agentbox/backend/internal/controller/image"
	"john-ai-agentbox/backend/internal/controller/middleware"
	promptcontroller "john-ai-agentbox/backend/internal/controller/prompt"
	providercontroller "john-ai-agentbox/backend/internal/controller/provider"
	serviceproxycontroller "john-ai-agentbox/backend/internal/controller/serviceproxy"
	settingcontroller "john-ai-agentbox/backend/internal/controller/setting"
	workspacecontroller "john-ai-agentbox/backend/internal/controller/workspace"
	accesssvc "john-ai-agentbox/backend/internal/service/access"
	aisvc "john-ai-agentbox/backend/internal/service/ai"
	authsvc "john-ai-agentbox/backend/internal/service/auth"
	catalogsvc "john-ai-agentbox/backend/internal/service/catalog"
	chatsvc "john-ai-agentbox/backend/internal/service/chat"
	containersvc "john-ai-agentbox/backend/internal/service/container"
	gatewaysvc "john-ai-agentbox/backend/internal/service/gateway"
	promptsvc "john-ai-agentbox/backend/internal/service/prompt"
	serviceproxysvc "john-ai-agentbox/backend/internal/service/serviceproxy"
	settingsvc "john-ai-agentbox/backend/internal/service/setting"
	terminalsvc "john-ai-agentbox/backend/internal/service/terminal"
	workspacesvc "john-ai-agentbox/backend/internal/service/workspace"

	"lina-core/pkg/plugin/pluginhost"
)

// Register binds AgentBox backend routes under /x/john-ai-agentbox/api/v1.
func Register(ctx context.Context, registrar pluginhost.HTTPRegistrar, assets fs.FS) error {
	if registrar == nil || registrar.Routes() == nil {
		return nil
	}
	routeRegistrar := registrar.Routes()
	if err := registerPortalRoutes(routeRegistrar, assets); err != nil {
		return err
	}
	config, err := loadAgentBoxConfig(ctx, registrar)
	if err != nil {
		return err
	}
	authStore := authsvc.NewDAOStore()
	authSvc, err := authsvc.New(authStore, config.Auth)
	if err != nil {
		return err
	}
	catalogHTTPClient := &http.Client{Timeout: config.Catalog.RemoteRequestTimeout}
	dockerRuntimeBackend := containersvc.NewDockerRuntimeBackend(config.Docker)
	catalogConfig := config.Catalog
	catalogConfig.AgentRuntimeBackend = dockerRuntimeBackend
	catalogSvc, err := catalogsvc.New(catalogHTTPClient, catalogConfig)
	if err != nil {
		return err
	}
	aiSvc, err := aisvc.New(catalogHTTPClient, config.AI)
	if err != nil {
		return err
	}
	settingSvc, err := settingsvc.New()
	if err != nil {
		return err
	}
	accessSvc, err := accesssvc.New(accesssvc.NewDAOStore())
	if err != nil {
		return err
	}
	chatSvc, err := chatsvc.New(accessSvc)
	if err != nil {
		return err
	}
	terminalSvc, err := terminalsvc.New(accessSvc)
	if err != nil {
		return err
	}
	workspaceSvc, err := workspacesvc.New(accessSvc, dockerRuntimeBackend, config.Workspace)
	if err != nil {
		return err
	}
	serviceProxySvc, err := serviceproxysvc.New(accessSvc, dockerRuntimeBackend, config.Service)
	if err != nil {
		return err
	}
	gatewaySvc, err := gatewaysvc.New(accessSvc)
	if err != nil {
		return err
	}
	containerSvc, err := containersvc.New(dockerRuntimeBackend, dockerRuntimeBackend)
	if err != nil {
		return err
	}
	promptSvc, err := promptsvc.New(promptsvc.NewDAOStore())
	if err != nil {
		return err
	}
	authController := authcontroller.NewV1(authSvc)
	agentController, err := agentcontroller.NewV1(catalogSvc, chatSvc, terminalSvc)
	if err != nil {
		return err
	}
	aiController := aicontroller.NewV1(aiSvc)
	providerController := providercontroller.NewV1(catalogSvc)
	imageController := imagecontroller.NewV1(catalogSvc)
	settingController := settingcontroller.NewV1(settingSvc)
	promptController := promptcontroller.NewV1(promptSvc)
	workspaceController, err := workspacecontroller.NewV1(workspaceSvc)
	if err != nil {
		return err
	}
	serviceProxyController, err := serviceproxycontroller.NewV1(serviceProxySvc)
	if err != nil {
		return err
	}
	containerController, err := containercontroller.NewV1(containerSvc)
	if err != nil {
		return err
	}
	gatewayController, err := gatewaycontroller.New(gatewaySvc)
	if err != nil {
		return err
	}
	routeRegistrar.Group(routeRegistrar.APIPrefix()+"/api/v1", func(group pluginhost.RouteGroup) {
		if middlewares := routeRegistrar.Middlewares(); middlewares != nil {
			group.Middleware(middlewares.HandlerResponse())
		}
		group.Bind(authController)
		group.Group("/", func(protected pluginhost.RouteGroup) {
			protected.Middleware(middleware.Auth(authSvc))
			protected.Bind(agentController, aiController, providerController, imageController, settingController, promptController, workspaceController, serviceProxyController, containerController)
			protected.ALL("/proxy/*", gatewayController.AgentServiceHTTPProxy)
			protected.GET("/ws/agents/{id}/shell", gatewayController.AgentShell)
			protected.GET("/ws/agents/{id}/chat/sessions/{sessionId}", gatewayController.AgentChat)
			protected.GET("/ws/agents/{id}/services/{serviceId}/tcp", gatewayController.AgentServiceTCPTunnel)
		})
	})
	return routeRegistrar.Err()
}
