// This file exposes hosted frontend asset methods on the root plugin facade.

package plugin

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
)

// PrewarmRuntimeFrontendBundles preloads frontend bundles for enabled dynamic plugins.
func (s *serviceImpl) PrewarmRuntimeFrontendBundles(ctx context.Context) error {
	readCtx, err := s.catalogSvc.WithStartupDataSnapshot(ctx)
	if err != nil {
		return err
	}
	return s.frontendSvc.PrewarmRuntimeFrontendBundles(readCtx)
}

// ResolveRuntimeFrontendAsset resolves one frontend asset for a dynamic plugin.
func (s *serviceImpl) ResolveRuntimeFrontendAsset(
	ctx context.Context,
	pluginID string,
	version string,
	relativePath string,
) (*RuntimeFrontendAssetOutput, error) {
	if err := s.ensureRuntimeCacheFresh(ctx); err != nil {
		return nil, err
	}
	if !s.integrationSvc.IsInstalledEnabledForTenant(ctx, pluginID) {
		return nil, errPluginPublicAssetNotFound(pluginID)
	}
	return s.frontendSvc.ResolveRuntimeFrontendAsset(ctx, pluginID, version, relativePath)
}

// BuildRuntimeFrontendPublicBaseURL returns the public base URL for a plugin's hosted frontend assets.
func (s *serviceImpl) BuildRuntimeFrontendPublicBaseURL(pluginID string, version string) string {
	return s.frontendSvc.BuildRuntimeFrontendPublicBaseURL(pluginID, version)
}

// errPluginPublicAssetNotFound returns a deliberately generic not-found error
// so disabled or tenant-unavailable plugins do not reveal public asset presence.
func errPluginPublicAssetNotFound(pluginID string) error {
	return gerror.Newf("plugin public asset is not available: %s", pluginID)
}
