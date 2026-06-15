// This file registers AgentBox browser entry points without requiring host
// portal fallback support. The entry points serve the plugin-owned SPA index
// directly while static assets continue to use the public asset namespace.

package routes

import (
	"io/fs"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/plugin/pluginhost"
)

const (
	portalIndexAssetPath  = "frontend/public/index.html"
	portalHTMLContentType = "text/html; charset=utf-8"
)

// registerPortalRoutes binds only exact browser entry points owned by AgentBox.
// It intentionally avoids wildcard fallback so host APIs, plugin APIs, plugin
// assets, and the management workbench remain governed by the host.
func registerPortalRoutes(routeRegistrar pluginhost.RouteRegistrar, assets fs.FS) error {
	if routeRegistrar == nil {
		return nil
	}
	handler, err := newPortalIndexHandler(assets)
	if err != nil {
		return err
	}
	routeRegistrar.Group("/", func(group pluginhost.RouteGroup) {
		group.GET("/", handler)
		group.GET("/login", handler)
	})
	return nil
}

func newPortalIndexHandler(assets fs.FS) (ghttp.HandlerFunc, error) {
	if assets == nil {
		return nil, gerror.New("agentbox portal asset filesystem is required")
	}
	content, err := fs.ReadFile(assets, portalIndexAssetPath)
	if err != nil {
		return nil, gerror.Wrapf(err, "read AgentBox portal index %s", portalIndexAssetPath)
	}
	if len(content) == 0 {
		return nil, gerror.Newf("AgentBox portal index %s is empty", portalIndexAssetPath)
	}
	indexHTML := append([]byte(nil), content...)
	return func(r *ghttp.Request) {
		r.Response.Header().Set("Content-Type", portalHTMLContentType)
		r.Response.Write(indexHTML)
		r.ExitAll()
	}, nil
}

func isAgentBoxPortalRouteBinding(path string) bool {
	normalized := "/" + strings.Trim(strings.TrimSpace(path), "/")
	return normalized == "/" || normalized == "/login"
}
