// Package pluginhost defines the public backend extension contracts that source
// plugins use to register routes, hooks, cron jobs, and governance callbacks
// through grouped facade interfaces.
package pluginhost

import (
	`io/fs`

	`lina-core/pkg/pluginservice/contract`
)

const (
	// PluginAPINamespaceSegment is the first URL path segment reserved for plugin APIs.
	PluginAPINamespaceSegment = "x"
	// PluginAPINamespacePrefix is the public URL prefix reserved for plugin APIs.
	PluginAPINamespacePrefix = "/" + PluginAPINamespaceSegment
	// HostedAssetPathSegment is the first URL path segment for host-served plugin public assets.
	HostedAssetPathSegment = "x-assets"
	// HostedAssetURLPrefix is the public URL prefix for host-served plugin public assets.
	HostedAssetURLPrefix = "/" + HostedAssetPathSegment + "/"
	// DynamicPageComponentPath is the workbench component used by dynamic plugin pages.
	DynamicPageComponentPath = "system/plugin/dynamic-page"
	// DynamicEmbeddedSourceQueryKey is the menu query key carrying an embedded asset URL.
	DynamicEmbeddedSourceQueryKey = "embeddedSrc"
	// DynamicAccessModeQueryKey is the menu query key controlling dynamic plugin page access mode.
	DynamicAccessModeQueryKey = "pluginAccessMode"
	// DynamicAccessModeEmbeddedMount is the access mode for ESM-mounted dynamic plugin pages.
	DynamicAccessModeEmbeddedMount = "embedded-mount"
)

// SourcePlugin defines the grouped plugin-facing contract published to source
// plugins during compile-time registration.
type SourcePlugin interface {
	// ID returns the stable plugin identifier that must match `plugin.yaml`.
	ID() string
	// Assets returns the plugin asset registration facade.
	Assets() SourcePluginAssets
	// Lifecycle returns the plugin lifecycle callback registration facade.
	Lifecycle() SourcePluginLifecycle
	// Hooks returns the event-hook registration facade.
	Hooks() SourcePluginHooks
	// HTTP returns the HTTP registration facade.
	HTTP() SourcePluginHTTP
	// Cron returns the cron registration facade.
	Cron() SourcePluginCron
	// Governance returns the menu and permission governance registration facade.
	Governance() SourcePluginGovernance
}

// HostServices exposes host-owned pluginservice adapters to source plugins.
type HostServices interface {
	// APIDoc returns the host API-documentation localization adapter used by
	// source plugins to resolve stable OpenAPI operation keys into localized
	// module titles and operation summaries. The method itself returns no error;
	// implementations may return nil only when the host service directory is not
	// initialized, and callers must treat the returned adapter as read-only host
	// catalog access.
	//
	// APIDoc 返回宿主 API 文档本地化适配器，供源码插件将稳定的 OpenAPI 操作键解析为本地化模块标题和操作摘要。
	// 该方法本身不返回错误；仅当宿主服务目录未初始化时实现才可以返回 nil，调用方必须把返回的适配器视为只读的宿主目录访问能力。
	APIDoc() contract.APIDocService

	// Auth returns the host tenant-auth adapter used by source plugins to
	// delegate tenant selection, tenant switching, and governed impersonation
	// token lifecycle operations back to the host. The returned service owns
	// token signing, session registration, revocation, and permission-cache
	// priming; plugins must still perform their own business authorization before
	// requesting privileged token operations.
	//
	// Auth 返回宿主租户认证适配器，供源码插件把租户选择、租户切换以及受治理的模拟登录令牌生命周期操作委托给宿主。
	// 返回服务负责令牌签发、会话注册、撤销和权限缓存预热；插件在请求高权限令牌操作前仍必须先完成自身业务授权判断。
	Auth() contract.AuthService

	// BizCtx returns the host business-context adapter that exposes a stable,
	// read-only projection of request identity, tenant, platform-bypass, and
	// impersonation metadata. Plugins should use this adapter instead of host
	// internal context types; absent request context fields are represented by
	// zero values in the returned contract model.
	//
	// BizCtx 返回宿主业务上下文适配器，用于暴露稳定、只读的请求身份、租户、平台绕过和模拟登录元数据投影。
	// 插件应使用该适配器而不是宿主内部上下文类型；请求上下文字段不存在时会在契约模型中体现为零值。
	BizCtx() contract.BizCtxService

	// Cache returns the plugin-scoped host cache adapter for transient,
	// tenant-aware plugin runtime data. The adapter binds cache access to the
	// current plugin identity and host cache backend; an unscoped base directory
	// may return nil because cache reads and writes require a plugin-bound service
	// view.
	//
	// Cache 返回插件作用域的宿主缓存适配器，用于存取具备租户感知能力的插件临时运行时数据。
	// 该适配器会把缓存访问绑定到当前插件身份和宿主缓存后端；未绑定插件的基础目录可以返回 nil，
	// 因为缓存读写必须使用插件绑定后的服务视图。
	Cache() contract.CacheService

	// Config returns the plugin-scoped static configuration adapter for reading
	// host-approved plugin configuration values. The adapter must not fall back to
	// unrestricted host-wide configuration when plugin identity is blank; an
	// unscoped base directory may return nil until HostServicesForPlugin binds a
	// plugin ID.
	//
	// Config 返回插件作用域的静态配置适配器，用于读取经过宿主认可的插件配置值。
	// 当插件身份为空时，该适配器不得回退读取不受限制的宿主全局配置；
	// 未绑定插件的基础目录可以在 HostServicesForPlugin 绑定插件 ID 前返回 nil。
	Config() contract.ConfigService

	// HostConfig returns the public host config adapter for whitelisted
	// configuration values that plugins may read without depending on private host
	// configuration models. Missing or unavailable keys are handled by the
	// returned service contract, commonly through default-value reads or errors
	// from typed accessors.
	//
	// HostConfig 返回宿主公开配置适配器，用于读取白名单内、允许插件访问的宿主配置值，避免插件依赖宿主私有配置模型。
	// 键不存在或不可用时由返回服务的契约处理，通常通过类型化读取方法返回默认值或错误。
	HostConfig() contract.HostConfigService

	// I18n returns the host runtime translation adapter for resolving the current
	// request locale, runtime message keys, fallback text, and localized keyword
	// searches. The returned service is read-only and must preserve the host
	// locale resolution rules instead of letting plugins manage translation
	// caches directly.
	//
	// I18n 返回宿主运行时翻译适配器，用于解析当前请求语言、运行时消息键、兜底文本以及本地化关键词搜索。
	// 返回服务是只读能力，并且必须保持宿主语言解析规则，插件不得通过该能力直接管理翻译缓存。
	I18n() contract.I18nService

	// Manifest returns the plugin-scoped manifest resource adapter for read-only
	// access to declaration resources under the plugin manifest root. Paths are
	// resolved relative to the plugin manifest boundary; an unscoped base
	// directory may return nil because manifest reads require a plugin-bound
	// service view.
	//
	// Manifest 返回插件作用域的清单资源适配器，用于只读访问插件 manifest 根目录下的声明资源。
	// 路径会在插件清单边界内按相对路径解析；未绑定插件的基础目录可以返回 nil，因为清单读取必须使用插件绑定后的服务视图。
	Manifest() contract.ManifestService

	// Notify returns the host notification adapter used by source plugins to
	// publish governed messages into the host inbox pipeline or remove messages
	// by declared business source. The adapter owns host delivery records and
	// fan-out behavior; plugins provide source identifiers, message content, and
	// category intent through the contract input models.
	//
	// Notify 返回宿主通知适配器，供源码插件向宿主站内信管道发布受治理的消息，或按声明的业务来源删除消息。
	// 该适配器负责宿主投递记录和分发行为；插件通过契约输入模型提供来源标识、消息内容和分类意图。
	Notify() contract.NotifyService

	// PluginState returns the host plugin enablement adapter for checking whether
	// a plugin is installed, enabled, and visible for the current request scope.
	// The returned service may expose both snapshot-backed fast reads and
	// authoritative persisted-state reads; callers choose the method according to
	// the freshness requirements of their control path.
	//
	// PluginState 返回宿主插件启用状态适配器，用于判断插件在当前请求范围内是否已安装、已启用且可见。
	// 返回服务可以同时提供基于快照的快速读取和基于持久化状态的权威读取；调用方应按控制路径的新鲜度要求选择具体方法。
	PluginState() contract.PluginStateService

	// PluginLifecycle returns the host plugin lifecycle orchestration adapter for
	// tenant-scoped plugin disable and tenant deletion coordination. The returned
	// service runs precondition checks that may return errors before destructive
	// governance actions and also publishes best-effort post-action
	// notifications after those actions complete.
	//
	// PluginLifecycle 返回宿主插件生命周期编排适配器，用于协调租户范围的插件禁用和租户删除流程。
	// 返回服务会在破坏性治理动作前执行可能返回错误的前置检查，并在动作完成后发布尽力而为的后置通知。
	PluginLifecycle() contract.PluginLifecycleService

	// Route returns the host dynamic-route metadata adapter for reading metadata
	// attached to the current dynamic-plugin HTTP request. It exposes the matched
	// plugin ID, method, public path, route text, metadata, and captured bridge
	// response details without exposing the host router internals.
	//
	// Route 返回宿主动态路由元数据适配器，用于读取当前动态插件 HTTP 请求上附加的路由元数据。
	// 它会暴露命中的插件 ID、方法、公开路径、路由文案、元数据以及捕获的桥接响应信息，但不会暴露宿主路由器内部实现。
	Route() contract.RouteService

	// Session returns the host online-session adapter for governed session list
	// and revocation operations. The returned service projects stable session
	// fields for plugins while preserving host authorization, tenant visibility,
	// filtering, pagination, and revocation semantics.
	//
	// Session 返回宿主在线会话适配器，用于受治理的会话列表查询和会话撤销操作。
	// 返回服务向插件投影稳定的会话字段，同时保持宿主授权、租户可见性、过滤、分页和撤销语义。
	Session() contract.SessionService

	// TenantFilter returns the host tenant-filter adapter for applying the
	// conventional tenant predicate to plugin-owned database queries and for
	// reading tenant/audit identity metadata. The adapter centralizes platform
	// bypass and impersonation decisions so plugins do not duplicate host tenant
	// visibility rules.
	//
	// TenantFilter 返回宿主租户过滤适配器，用于向插件自有数据库查询应用约定的租户谓词，并读取租户和审计身份元数据。
	// 该适配器集中处理平台绕过和模拟登录判断，避免插件重复实现宿主租户可见性规则。
	TenantFilter() contract.TenantFilterService
}

// SourcePluginAssets exposes plugin-owned asset declarations grouped under one
// dedicated facade.
type SourcePluginAssets interface {
	// UseEmbeddedFiles binds one plugin-owned embedded filesystem.
	UseEmbeddedFiles(fileSystem fs.FS)
}

// SourcePluginLifecycle exposes lifecycle callback registrations grouped under
// one dedicated facade.
type SourcePluginLifecycle interface {
	// RegisterBeforeInstallHandler registers a pre-install lifecycle callback
	// for the source plugin. The host invokes this callback before it applies
	// install SQL, synchronizes plugin governance resources, or marks the plugin
	// as installed. Return ok=false to veto installation with a stable reason
	// key, or return an error when the precondition check itself failed. Use this
	// hook when installation depends on external configuration, tenant readiness,
	// license state, host capability checks, or other conditions that must be
	// satisfied before any install side effects are written.
	RegisterBeforeInstallHandler(handler SourcePluginBeforeLifecycleHandler) error
	// RegisterAfterInstallHandler registers a post-install lifecycle callback
	// for the source plugin. The host invokes this callback after install SQL,
	// governance synchronization, registry state update, release synchronization,
	// metadata synchronization, and cache refresh signals have completed. Use it
	// for follow-up work that observes a successful install, such as warming
	// plugin-local caches, emitting telemetry, or scheduling asynchronous
	// reconciliation. A failure is logged by the host and does not roll back the
	// already-effective installation.
	RegisterAfterInstallHandler(handler SourcePluginAfterLifecycleHandler) error
	// RegisterBeforeUpgradeHandler registers a pre-upgrade lifecycle callback
	// for the source plugin. The host invokes this callback after it has built
	// the upgrade plan and before it runs the plugin's custom upgrade handler,
	// upgrade SQL, governance synchronization, release switch, or cache
	// invalidation. Return ok=false to stop the upgrade with a stable reason key.
	// Use this hook to validate compatibility between the effective manifest and
	// the discovered target manifest, block unsupported version jumps, verify
	// required host services, or enforce plugin-specific migration prerequisites.
	RegisterBeforeUpgradeHandler(handler SourcePluginBeforeUpgradeHandler) error
	// RegisterUpgradeHandler registers the plugin-owned upgrade callback that
	// runs during a source-plugin runtime upgrade. The host invokes this callback
	// after all pre-upgrade callbacks allow the operation and before it executes
	// upgrade SQL and promotes the target release. Use this hook for custom,
	// version-aware migration work that cannot be represented by manifest SQL
	// alone, such as transforming plugin-owned data, preparing external
	// resources, or bridging data between old and new plugin contracts. The
	// callback should be idempotent or safely retryable because a failed upgrade
	// can be retried by an operator.
	RegisterUpgradeHandler(handler SourcePluginUpgradeHandler) error
	// RegisterAfterUpgradeHandler registers a post-upgrade lifecycle callback
	// for the source plugin. The host invokes this callback after upgrade SQL,
	// governance synchronization, release promotion, node-state synchronization,
	// and cache refresh signals have completed successfully. Use this hook for
	// best-effort follow-up work that observes the new effective version, such
	// as warming plugin caches, emitting plugin-local telemetry, refreshing
	// external integrations, or scheduling asynchronous reconciliation. A failure
	// is logged by the host and does not roll back the already-effective upgrade.
	RegisterAfterUpgradeHandler(handler SourcePluginUpgradeHandler) error
	// RegisterBeforeDisableHandler registers a pre-disable lifecycle callback
	// for the source plugin. The host invokes this callback before changing the
	// plugin from enabled to disabled and before business entry points are hidden
	// or stopped. Return ok=false to veto the disable operation with a stable
	// reason key. Use this hook when the plugin must prevent disable while jobs,
	// workflows, external subscriptions, tenant obligations, or other
	// plugin-owned runtime work is still active.
	RegisterBeforeDisableHandler(handler SourcePluginBeforeLifecycleHandler) error
	// RegisterAfterDisableHandler registers a post-disable lifecycle callback
	// for the source plugin. The host invokes this callback after the plugin has
	// been disabled, business entry points have been hidden or stopped, cache
	// refresh signals have completed, and lifecycle observers have been notified.
	// Use it for best-effort follow-up work such as closing external sessions,
	// emitting telemetry, or scheduling reconciliation. A failure is logged by
	// the host and does not roll back the disable operation.
	RegisterAfterDisableHandler(handler SourcePluginAfterLifecycleHandler) error
	// RegisterBeforeUninstallHandler registers a pre-uninstall lifecycle callback
	// for the source plugin. The host invokes this callback before it runs
	// plugin cleanup, uninstall SQL, governance resource deletion, registry state
	// changes, and uninstall hook events. Return ok=false to veto normal
	// uninstall with a stable reason key; force uninstall may bypass the veto
	// only when the host configuration explicitly permits it. Use this hook to
	// protect plugin-owned data, block uninstall while dependent resources still
	// exist, require operator confirmation outside the host, or verify that
	// external cleanup prerequisites are satisfied.
	RegisterBeforeUninstallHandler(handler SourcePluginBeforeLifecycleHandler) error
	// RegisterAfterUninstallHandler registers a post-uninstall lifecycle callback
	// for the source plugin. The host invokes this callback after plugin cleanup,
	// uninstall SQL when requested, governance deletion, registry state update,
	// release synchronization, metadata synchronization, cache refresh signals,
	// and lifecycle observers have completed. Use it for best-effort telemetry or
	// external reconciliation that should observe the final uninstalled state. A
	// failure is logged by the host and does not roll back uninstall.
	RegisterAfterUninstallHandler(handler SourcePluginAfterLifecycleHandler) error
	// RegisterBeforeTenantDisableHandler registers a tenant-scoped pre-disable
	// lifecycle callback for the source plugin. The host invokes this callback
	// before disabling the plugin for one tenant while leaving global plugin
	// installation state intact. Return ok=false to veto the tenant-scoped
	// disable with a stable reason key. Use this hook when tenant-specific
	// plugin activity, subscriptions, pending work, or data retention policy
	// must be checked before removing that tenant's access to the plugin.
	RegisterBeforeTenantDisableHandler(handler SourcePluginBeforeTenantLifecycleHandler) error
	// RegisterAfterTenantDisableHandler registers a tenant-scoped post-disable
	// lifecycle callback for the source plugin. The host invokes this callback
	// after one tenant has successfully lost access to the plugin. Use it for
	// tenant-local cache warming, telemetry, or external reconciliation. A
	// failure is logged by the host and does not roll back tenant disable.
	RegisterAfterTenantDisableHandler(handler SourcePluginAfterTenantLifecycleHandler) error
	// RegisterBeforeTenantDeleteHandler registers a tenant-delete precondition
	// callback for the source plugin. The host invokes this callback before a
	// tenant is deleted so installed plugins can protect tenant-owned plugin
	// data and external resources. Return ok=false to block tenant deletion with
	// a stable reason key. Use this hook when the plugin stores tenant-scoped
	// records, owns external tenant mappings, runs tenant-specific jobs, or must
	// require explicit cleanup before the tenant can be removed.
	RegisterBeforeTenantDeleteHandler(handler SourcePluginBeforeTenantLifecycleHandler) error
	// RegisterAfterTenantDeleteHandler registers a tenant-delete post-notification
	// callback for the source plugin. The host invokes this callback after the
	// tenant has been deleted and plugin-owned preconditions have passed. Use it
	// for best-effort cleanup of external tenant mappings or telemetry. A failure
	// is logged by the host and does not roll back tenant deletion.
	RegisterAfterTenantDeleteHandler(handler SourcePluginAfterTenantLifecycleHandler) error
	// RegisterBeforeInstallModeChangeHandler registers a precondition callback
	// for source-plugin install-mode transitions. The host invokes this callback
	// before switching the plugin between supported install modes, such as
	// global and tenant-scoped modes. Return ok=false to veto the transition with
	// a stable reason key. Use this hook when a mode change would alter tenant
	// visibility, data ownership, governance resources, or runtime assumptions
	// and the plugin must verify that existing data and active tenants can be
	// migrated or safely preserved.
	RegisterBeforeInstallModeChangeHandler(handler SourcePluginBeforeInstallModeChangeHandler) error
	// RegisterAfterInstallModeChangeHandler registers a post-notification
	// callback for source-plugin install-mode transitions. The host invokes this
	// callback after an install-mode transition succeeds. Use it for follow-up
	// reconciliation, telemetry, or cache warming that observes the new mode. A
	// failure is logged by the host and does not roll back the mode change.
	RegisterAfterInstallModeChangeHandler(handler SourcePluginAfterInstallModeChangeHandler) error
	// RegisterUninstallHandler registers the plugin-owned cleanup callback that
	// runs during uninstall when the operator requested storage/data purging. The
	// host invokes this callback after uninstall preconditions have passed and
	// before uninstall SQL removes plugin-owned tables. Use this hook to delete
	// or detach external resources, remove plugin-managed files, revoke external
	// subscriptions, or perform cleanup that cannot be expressed as uninstall
	// SQL. The callback should be idempotent because uninstall can be retried
	// after cleanup or SQL failures.
	RegisterUninstallHandler(handler SourcePluginUninstallHandler) error
}

// SourcePluginHooks exposes callback-style host hook registrations grouped
// under one dedicated facade.
type SourcePluginHooks interface {
	// RegisterHook registers one callback-style host hook handler.
	RegisterHook(point ExtensionPoint, mode CallbackExecutionMode, handler HookHandler) error
}

// SourcePluginHTTP exposes HTTP-adjacent registrations grouped under one
// dedicated facade.
type SourcePluginHTTP interface {
	// RegisterRoutes registers one callback that contributes plugin-owned HTTP routes.
	RegisterRoutes(point ExtensionPoint, mode CallbackExecutionMode, handler RouteRegisterHandler) error
}

// SourcePluginCron exposes cron registrations grouped under one dedicated
// facade.
type SourcePluginCron interface {
	// RegisterCron registers one callback that contributes plugin-owned cron jobs.
	RegisterCron(point ExtensionPoint, mode CallbackExecutionMode, handler CronRegisterHandler) error
}

// SourcePluginGovernance exposes governance callback registrations grouped
// under one dedicated facade.
type SourcePluginGovernance interface {
	// RegisterMenuFilter registers one callback that filters host menus.
	RegisterMenuFilter(point ExtensionPoint, mode CallbackExecutionMode, handler MenuFilterHandler) error
	// RegisterPermissionFilter registers one callback that filters host permissions.
	RegisterPermissionFilter(point ExtensionPoint, mode CallbackExecutionMode, handler PermissionFilterHandler) error
}

// SourcePluginDefinition exposes the host-side read model restored from one
// grouped source-plugin registration.
type SourcePluginDefinition interface {
	SourcePlugin
	// GetEmbeddedFiles returns the plugin-owned embedded filesystem when declared.
	GetEmbeddedFiles() fs.FS
	// GetHookHandlers returns the registered callback-style hook handlers.
	GetHookHandlers() []*HookHandlerRegistration
	// GetRouteRegistrars returns the registered route contribution callbacks.
	GetRouteRegistrars() []*RouteHandlerRegistration
	// GetCronRegistrars returns the registered cron contribution callbacks.
	GetCronRegistrars() []*CronHandlerRegistration
	// GetMenuFilters returns the registered menu filter callbacks.
	GetMenuFilters() []*MenuFilterHandlerRegistration
	// GetPermissionFilters returns the registered permission filter callbacks.
	GetPermissionFilters() []*PermissionFilterHandlerRegistration
	// GetBeforeInstallHandler returns the registered pre-install veto callback.
	GetBeforeInstallHandler() SourcePluginBeforeLifecycleHandler
	// GetAfterInstallHandler returns the registered post-install callback.
	GetAfterInstallHandler() SourcePluginAfterLifecycleHandler
	// GetBeforeUpgradeHandler returns the registered pre-upgrade veto callback.
	GetBeforeUpgradeHandler() SourcePluginBeforeUpgradeHandler
	// GetUpgradeHandler returns the registered source-plugin custom upgrade callback.
	GetUpgradeHandler() SourcePluginUpgradeHandler
	// GetAfterUpgradeHandler returns the registered post-upgrade event callback.
	GetAfterUpgradeHandler() SourcePluginUpgradeHandler
	// GetBeforeDisableHandler returns the registered pre-disable veto callback.
	GetBeforeDisableHandler() SourcePluginBeforeLifecycleHandler
	// GetAfterDisableHandler returns the registered post-disable callback.
	GetAfterDisableHandler() SourcePluginAfterLifecycleHandler
	// GetBeforeUninstallHandler returns the registered pre-uninstall veto callback.
	GetBeforeUninstallHandler() SourcePluginBeforeLifecycleHandler
	// GetAfterUninstallHandler returns the registered post-uninstall callback.
	GetAfterUninstallHandler() SourcePluginAfterLifecycleHandler
	// GetBeforeTenantDisableHandler returns the registered tenant-disable veto callback.
	GetBeforeTenantDisableHandler() SourcePluginBeforeTenantLifecycleHandler
	// GetAfterTenantDisableHandler returns the registered tenant-disable post callback.
	GetAfterTenantDisableHandler() SourcePluginAfterTenantLifecycleHandler
	// GetBeforeTenantDeleteHandler returns the registered tenant-delete veto callback.
	GetBeforeTenantDeleteHandler() SourcePluginBeforeTenantLifecycleHandler
	// GetAfterTenantDeleteHandler returns the registered tenant-delete post callback.
	GetAfterTenantDeleteHandler() SourcePluginAfterTenantLifecycleHandler
	// GetBeforeInstallModeChangeHandler returns the registered install-mode change veto callback.
	GetBeforeInstallModeChangeHandler() SourcePluginBeforeInstallModeChangeHandler
	// GetAfterInstallModeChangeHandler returns the registered install-mode change post callback.
	GetAfterInstallModeChangeHandler() SourcePluginAfterInstallModeChangeHandler
	// GetUninstallHandler returns the registered source-plugin uninstall cleanup callback.
	GetUninstallHandler() SourcePluginUninstallHandler
}
