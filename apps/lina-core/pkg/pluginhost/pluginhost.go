// Package pluginhost defines the public backend extension contracts that source
// plugins use to register routes, hooks, cron jobs, and governance callbacks
// through grouped facade interfaces.
package pluginhost

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
