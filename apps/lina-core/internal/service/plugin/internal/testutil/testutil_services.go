// This file wires plugin sub-services and request-context adapters for tests.

package testutil

import (
	"context"
	"fmt"
	"time"

	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	configsvc "lina-core/internal/service/config"
	"lina-core/internal/service/hostlock"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/kvcache"
	"lina-core/internal/service/locker"
	"lina-core/internal/service/notify"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/frontend"
	"lina-core/internal/service/plugin/internal/integration"
	"lina-core/internal/service/plugin/internal/lifecycle"
	"lina-core/internal/service/plugin/internal/openapi"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/wasm"
	tenantcapsvc "lina-core/internal/service/tenantcap"
	"lina-core/pkg/pluginhost"
	pluginserviceconfig "lina-core/pkg/pluginservice/config"
	"lina-core/pkg/pluginservice/contract"
	pluginservicehostconfig "lina-core/pkg/pluginservice/hostconfig"
	pluginservicemanifest "lina-core/pkg/pluginservice/manifest"
)

// Services groups the wired plugin sub-services used by package-level tests.
type Services struct {
	// Catalog provides manifest discovery, registry, and release access.
	Catalog catalog.Service
	// Lifecycle provides install and uninstall orchestration.
	Lifecycle lifecycle.Service
	// Runtime provides artifact parsing, reconcile, and route execution.
	Runtime runtime.Service
	// Frontend provides in-memory frontend bundle management.
	Frontend frontend.Service
	// Integration provides menu, hook, and resource-ref integration.
	Integration integration.Service
	// OpenAPI provides dynamic route OpenAPI projection.
	OpenAPI openapi.Service
}

// singleNodeTopology provides the default non-clustered topology for plugin tests.
type singleNodeTopology struct{}

// IsClusterModeEnabled reports that package tests run in single-node mode.
func (singleNodeTopology) IsClusterModeEnabled() bool {
	return false
}

// IsPrimaryNode reports that the local test node owns primary-only work.
func (singleNodeTopology) IsPrimaryNode() bool {
	return true
}

// CurrentNodeID returns the fixed node identifier used by package tests.
func (singleNodeTopology) CurrentNodeID() string {
	return "test-node"
}

// NewServices creates a fully wired plugin sub-service set for tests.
func NewServices() *Services {
	var (
		configProvider = configsvc.New()
		bizCtxProvider = bizctx.New()
		cacheCoordSvc  = cachecoord.Default(cachecoord.NewStaticTopology(false))
		i18nService    = i18nsvc.New(bizCtxProvider, configProvider, cacheCoordSvc)
		catalogSvc     = catalog.New(configProvider)
		lifecycleSvc   = lifecycle.New(catalogSvc)
		frontendSvc    = frontend.New(catalogSvc)
		openapiSvc     = openapi.New(catalogSvc)
		runtimeSvc     = runtime.New(catalogSvc, lifecycleSvc, frontendSvc, openapiSvc, i18nService)
		integrationSvc = integration.New(catalogSvc)
		topology       = singleNodeTopology{}
		kvCacheSvc     = kvcache.New()
		tenantSvc      = tenantcapsvc.New(nil, bizCtxProvider)
		notifySvc      = notify.New(tenantSvc)
	)
	hostLockSvc := mustNewHostLockServiceForTest()

	catalogSvc.SetBackendLoader(integrationSvc)
	catalogSvc.SetArtifactParser(runtimeSvc)
	catalogSvc.SetDynamicManifestLoader(runtimeSvc)
	catalogSvc.SetNodeStateSyncer(runtimeSvc)
	catalogSvc.SetMenuSyncer(integrationSvc)
	catalogSvc.SetResourceRefSyncer(integrationSvc)
	catalogSvc.SetReleaseStateSyncer(runtimeSvc)
	catalogSvc.SetHookDispatcher(integrationSvc)

	lifecycleSvc.SetReconciler(runtimeSvc)
	lifecycleSvc.SetTopology(topology)

	integrationSvc.SetBizCtxProvider(&bizCtxAdapter{svc: bizCtxProvider})
	integrationSvc.SetDynamicCronExecutor(runtimeSvc)
	integrationSvc.SetHostServices(newTestHostServices())
	integrationSvc.SetTopologyProvider(topology)

	runtimeSvc.SetMenuManager(integrationSvc)
	runtimeSvc.SetHookDispatcher(integrationSvc)
	runtimeSvc.SetPermissionMenuFilter(integrationSvc)
	runtimeSvc.SetJwtConfigProvider(&jwtConfigAdapter{svc: configProvider})
	runtimeSvc.SetUploadSizeProvider(&uploadSizeAdapter{svc: configProvider})
	runtimeSvc.SetUserContextSetter(&userCtxAdapter{svc: bizCtxProvider})
	runtimeSvc.SetTopology(topology)

	mustConfigureWasmHostServicesForTest(
		kvCacheSvc,
		hostLockSvc,
		notifySvc,
		configProvider,
		pluginserviceconfig.NewFactory("", ""),
		pluginservicehostconfig.New(configProvider),
		pluginservicemanifest.NewFactory(""),
	)

	return &Services{
		Catalog:     catalogSvc,
		Lifecycle:   lifecycleSvc,
		Runtime:     runtimeSvc,
		Frontend:    frontendSvc,
		Integration: integrationSvc,
		OpenAPI:     openapiSvc,
	}
}

// mustNewHostLockServiceForTest creates the host-lock dependency used by wasm
// bridge tests. A failure means the fixture wiring is invalid.
func mustNewHostLockServiceForTest() hostlock.Service {
	service, err := hostlock.New(locker.New())
	if err != nil {
		panic(fmt.Sprintf("configure test host lock service: %v", err))
	}
	return service
}

// mustConfigureWasmHostServicesForTest mirrors the HTTP startup host-service
// wiring so runtime package tests are self-contained and order independent.
func mustConfigureWasmHostServicesForTest(
	kvCacheSvc kvcache.Service,
	hostLockSvc hostlock.Service,
	notifySvc notify.Service,
	configProvider configsvc.Service,
	configFactory contract.ConfigServiceFactory,
	hostConfigSvc contract.HostConfigService,
	manifestFactory contract.ManifestServiceFactory,
) {
	configure := []struct {
		name string
		fn   func() error
	}{
		{name: "cache", fn: func() error { return wasm.ConfigureCacheHostService(kvCacheSvc) }},
		{name: "lock", fn: func() error { return wasm.ConfigureLockHostService(hostLockSvc) }},
		{name: "notify", fn: func() error { return wasm.ConfigureNotifyHostService(notifySvc) }},
		{name: "storage", fn: func() error { return wasm.ConfigureStorageHostService(configProvider) }},
		{name: "config", fn: func() error { return wasm.ConfigureConfigHostService(configFactory) }},
		{name: "host config", fn: func() error { return wasm.ConfigureHostConfigService(hostConfigSvc) }},
		{name: "manifest", fn: func() error { return wasm.ConfigureManifestHostService(manifestFactory) }},
	}
	for _, item := range configure {
		if err := item.fn(); err != nil {
			panic(fmt.Sprintf("configure test wasm %s host service: %v", item.name, err))
		}
	}
}

// testHostServices publishes the minimal host service directory needed by
// source-plugin callbacks exercised in plugin service tests.
type testHostServices struct {
	// configFactory creates plugin-scoped configuration views.
	configFactory contract.ConfigServiceFactory
	// manifestFactory creates plugin-scoped manifest resource views.
	manifestFactory contract.ManifestServiceFactory
	// pluginID scopes source-plugin host services when non-empty.
	pluginID string
}

// Ensure testHostServices satisfies the source-plugin host service directory.
var _ pluginhost.HostServices = (*testHostServices)(nil)

// Ensure testHostServices can return plugin-scoped host service views.
var _ pluginhost.ScopedHostServicesFactory = (*testHostServices)(nil)

// newTestHostServices creates a host service directory for integration tests.
func newTestHostServices() pluginhost.HostServices {
	return &testHostServices{
		configFactory:   pluginserviceconfig.NewFactory("", ""),
		manifestFactory: pluginservicemanifest.NewFactory(""),
	}
}

// APIDoc returns no apidoc service for plugin integration tests.
func (s *testHostServices) APIDoc() contract.APIDocService { return nil }

// Auth returns no auth service for plugin integration tests.
func (s *testHostServices) Auth() contract.AuthService { return nil }

// BizCtx returns no bizctx service for plugin integration tests.
func (s *testHostServices) BizCtx() contract.BizCtxService { return nil }

// Cache returns no cache service for plugin integration tests.
func (s *testHostServices) Cache() contract.CacheService { return nil }

// Config returns the plugin-scoped test host configuration service.
func (s *testHostServices) Config() contract.ConfigService {
	if s == nil || s.configFactory == nil {
		return nil
	}
	return s.configFactory.ForPlugin(s.pluginID)
}

// ForPlugin returns a plugin-bound host service view for source-plugin callbacks.
func (s *testHostServices) ForPlugin(pluginID string) pluginhost.HostServices {
	if s == nil {
		return nil
	}
	return &testHostServices{
		configFactory:   s.configFactory,
		manifestFactory: s.manifestFactory,
		pluginID:        pluginID,
	}
}

// HostConfig returns no host config service for plugin integration tests.
func (s *testHostServices) HostConfig() contract.HostConfigService { return nil }

// I18n returns no i18n service for plugin integration tests.
func (s *testHostServices) I18n() contract.I18nService { return nil }

// Manifest returns the plugin-scoped manifest service for plugin integration tests.
func (s *testHostServices) Manifest() contract.ManifestService {
	if s == nil || s.manifestFactory == nil {
		return nil
	}
	return s.manifestFactory.ForPlugin(s.pluginID)
}

// Notify returns no notification service for plugin integration tests.
func (s *testHostServices) Notify() contract.NotifyService { return nil }

// PluginLifecycle returns no lifecycle service for plugin integration tests.
func (s *testHostServices) PluginLifecycle() contract.PluginLifecycleService { return nil }

// PluginState returns no plugin-state service for plugin integration tests.
func (s *testHostServices) PluginState() contract.PluginStateService { return nil }

// Route returns no route service for plugin integration tests.
func (s *testHostServices) Route() contract.RouteService { return nil }

// Session returns no session service for plugin integration tests.
func (s *testHostServices) Session() contract.SessionService { return nil }

// TenantFilter returns no tenant-filter service for plugin integration tests.
func (s *testHostServices) TenantFilter() contract.TenantFilterService { return nil }

// jwtConfigAdapter exposes config service JWT settings through the runtime test seam.
type jwtConfigAdapter struct {
	// svc provides JWT runtime configuration.
	svc configsvc.Service
}

// GetJwtSecret returns the configured JWT signing secret for test wiring.
func (a *jwtConfigAdapter) GetJwtSecret(ctx context.Context) string {
	return a.svc.GetJwtSecret(ctx)
}

// GetSessionTimeout returns the runtime-effective session timeout for test wiring.
func (a *jwtConfigAdapter) GetSessionTimeout(ctx context.Context) (time.Duration, error) {
	return a.svc.GetSessionTimeout(ctx)
}

// uploadSizeAdapter exposes upload-size config through the runtime test seam.
type uploadSizeAdapter struct {
	// svc provides upload-size runtime configuration.
	svc configsvc.Service
}

// GetUploadMaxSize returns the runtime-effective upload limit used in tests.
func (a *uploadSizeAdapter) GetUploadMaxSize(ctx context.Context) (int64, error) {
	return a.svc.GetUploadMaxSize(ctx)
}

// userCtxAdapter forwards authenticated user injection to the shared bizctx service.
type userCtxAdapter struct {
	// svc stores request-local user context.
	svc bizctx.Service
}

// SetUser injects authenticated user identity into the test request context.
func (a *userCtxAdapter) SetUser(ctx context.Context, tokenID string, userID int, username string, status int) {
	a.svc.SetUser(ctx, tokenID, userID, username, status)
}

// SetTenant injects the resolved tenant into the test request context.
func (a *userCtxAdapter) SetTenant(ctx context.Context, tenantID int) {
	a.svc.SetTenant(ctx, tenantID)
}

// SetUserAccess injects cached access-snapshot fields into the test request context.
func (a *userCtxAdapter) SetUserAccess(ctx context.Context, dataScope int, dataScopeUnsupported bool, unsupportedDataScope int) {
	a.svc.SetUserAccess(ctx, dataScope, dataScopeUnsupported, unsupportedDataScope)
}

// bizCtxAdapter exposes the current request user ID to integration-layer tests.
type bizCtxAdapter struct {
	// svc reads request-local user context.
	svc bizctx.Service
}

// GetUserId returns the current request user ID for integration-layer tests.
func (a *bizCtxAdapter) GetUserId(ctx context.Context) int {
	localCtx := a.svc.Get(ctx)
	if localCtx == nil {
		return 0
	}
	return localCtx.UserId
}

// GetDataScope returns the current request user's effective role data-scope.
func (a *bizCtxAdapter) GetDataScope(ctx context.Context) int {
	localCtx := a.svc.Get(ctx)
	if localCtx == nil {
		return 0
	}
	return localCtx.DataScope
}

// GetDataScopeUnsupported returns the unsupported data-scope state from the current request.
func (a *bizCtxAdapter) GetDataScopeUnsupported(ctx context.Context) (bool, int) {
	localCtx := a.svc.Get(ctx)
	if localCtx == nil {
		return false, 0
	}
	return localCtx.DataScopeUnsupported, localCtx.UnsupportedDataScope
}
