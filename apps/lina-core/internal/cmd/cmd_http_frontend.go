// This file serves embedded frontend assets and dynamic plugin frontend assets.

package cmd

import (
	"context"
	"io/fs"
	"mime"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/packed"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/pkg/logger"
	"lina-core/pkg/pluginhost"
)

const frontendDevServerURLEnv = "LINAPRO_FRONTEND_DEV_SERVER_URL"

// bindFrontendAssetRoutes registers the final frontend catch-all route after
// API and plugin routes are bound. The handler only serves `/x-assets` and the
// configured workspace base path; other unmatched paths return 404.
func bindFrontendAssetRoutes(
	ctx context.Context,
	server *ghttp.Server,
	pluginSvc pluginsvc.Service,
	workspaceBasePath string,
) error {
	subFS, err := fs.Sub(packed.Files, "public")
	if err != nil {
		logger.Panicf(ctx, "load embedded frontend assets failed: %v", err)
		return err
	}
	fileServer := http.FileServer(http.FS(subFS))
	devProxy, err := newFrontendDevServerProxy()
	if err != nil {
		return err
	}
	normalizedWorkspaceBasePath := normalizeWorkspaceRequestBasePath(workspaceBasePath)
	assetHandler := newFrontendAssetHandler(subFS, fileServer, pluginSvc, devProxy, normalizedWorkspaceBasePath)
	server.BindHandler(pluginhost.HostedAssetURLPrefix, assetHandler)
	server.BindHandler(pluginhost.HostedAssetURLPrefix+"/*any", assetHandler)
	server.BindHandler("/"+normalizedWorkspaceBasePath, assetHandler)
	server.BindHandler("/"+normalizedWorkspaceBasePath+"/*any", assetHandler)
	server.BindHandler("/{entry}", assetHandler)
	server.BindHandler("/{entry}/*any", assetHandler)
	return nil
}

// newFrontendAssetHandler creates the guarded catch-all handler. It runs after
// host and source-plugin routes, so concrete plugin routes get first chance.
func newFrontendAssetHandler(
	subFS fs.FS,
	fileServer http.Handler,
	pluginSvc pluginsvc.Service,
	devProxy http.Handler,
	workspaceBasePath string,
) func(r *ghttp.Request) {
	return func(r *ghttp.Request) {
		requestPath := normalizeRequestPath(r.URL.Path)
		if serveRuntimePluginAsset(r, pluginSvc, requestPath) {
			return
		}
		workspacePath, ok := trimWorkspaceRequestPath(requestPath, workspaceBasePath)
		if !ok {
			r.Response.WriteStatus(http.StatusNotFound)
			r.ExitAll()
			return
		}
		if devProxy != nil {
			serveFrontendDevProxy(r, devProxy, requestPath, workspaceBasePath)
			r.ExitAll()
			return
		}
		if serveEmbeddedFrontendAsset(r, subFS, fileServer, workspacePath) {
			return
		}
		serveSPAFallback(r, fileServer)
	}
}

// serveFrontendDevProxy forwards workspace requests to Vite. Vite requires the
// configured base path to include its trailing slash, so exact `/admin` style
// requests are normalized before proxying.
func serveFrontendDevProxy(
	r *ghttp.Request,
	devProxy http.Handler,
	requestPath string,
	workspaceBasePath string,
) {
	proxyRequest := r.Request
	if strings.Trim(requestPath, "/") == strings.Trim(workspaceBasePath, "/") {
		proxyRequest = r.Request.Clone(r.Context())
		proxyRequest.URL.Path = "/" + strings.Trim(workspaceBasePath, "/") + "/"
		proxyRequest.URL.RawPath = ""
	}
	devProxy.ServeHTTP(r.Response.RawWriter(), proxyRequest)
}

// newFrontendDevServerProxy builds the optional development reverse proxy used
// by linactl dev. Production leaves the env unset and serves embedded assets.
func newFrontendDevServerProxy() (http.Handler, error) {
	rawURL := strings.TrimSpace(os.Getenv(frontendDevServerURLEnv))
	if rawURL == "" {
		return nil, nil
	}
	target, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	if target.Scheme != "http" && target.Scheme != "https" {
		return nil, url.InvalidHostError("frontend dev server URL must use http or https")
	}
	if strings.TrimSpace(target.Host) == "" {
		return nil, url.InvalidHostError("frontend dev server URL must include host")
	}
	return httputil.NewSingleHostReverseProxy(target), nil
}

// serveRuntimePluginAsset serves versioned dynamic plugin frontend assets when
// the request path belongs to the public plugin-asset namespace.
func serveRuntimePluginAsset(
	r *ghttp.Request,
	pluginSvc pluginsvc.Service,
	path string,
) bool {
	// Plugin public assets must be checked before the host falls back to the
	// embedded frontend bundle. They are governed by plugin ID, version,
	// public_assets declarations, enabled state, and tenant availability.
	pluginID, version, assetPath, ok := parsePluginAssetRequestPath(path)
	if !ok {
		return false
	}
	out, resolveErr := pluginSvc.ResolveRuntimeFrontendAsset(
		r.Context(),
		pluginID,
		version,
		assetPath,
	)
	if resolveErr != nil {
		r.Response.WriteStatus(http.StatusNotFound)
		r.ExitAll()
		return true
	}
	r.Response.Header().Set("Content-Type", out.ContentType)
	r.Response.Write(out.Content)
	r.ExitAll()
	return true
}

// serveEmbeddedFrontendAsset serves one concrete embedded frontend file when
// it exists and lets callers fall through to the SPA fallback otherwise.
func serveEmbeddedFrontendAsset(
	r *ghttp.Request,
	subFS fs.FS,
	fileServer http.Handler,
	assetPath string,
) bool {
	content, err := fs.ReadFile(subFS, assetPath)
	if err != nil {
		return false
	}
	contentType := mime.TypeByExtension(path.Ext(assetPath))
	if contentType == "" {
		contentType = http.DetectContentType(content)
	}
	r.Response.Header().Set("Content-Type", contentType)
	r.Response.Write(content)
	r.ExitAll()
	return true
}

// serveSPAFallback serves index.html for unmatched frontend routes so browser
// refreshes on client-side routes are handled by the Vue application.
func serveSPAFallback(r *ghttp.Request, fileServer http.Handler) {
	r.Request.URL.Path = "/index.html"
	fileServer.ServeHTTP(r.Response.RawWriter(), r.Request)
	r.ExitAll()
}

// normalizeRequestPath trims the leading slash while preserving sub-paths.
func normalizeRequestPath(rawPath string) string {
	return strings.TrimPrefix(strings.TrimSpace(rawPath), "/")
}

// normalizeWorkspaceRequestBasePath returns a slashless workspace base path for
// prefix checks against request paths.
func normalizeWorkspaceRequestBasePath(basePath string) string {
	normalized := strings.Trim(strings.TrimSpace(basePath), "/")
	if normalized == "" {
		return "admin"
	}
	return normalized
}

// trimWorkspaceRequestPath removes the workspace base path from an incoming
// request and returns the embedded-asset path that should be served.
func trimWorkspaceRequestPath(requestPath string, workspaceBasePath string) (string, bool) {
	normalizedRequestPath := strings.Trim(requestPath, "/")
	if normalizedRequestPath == "" {
		return "", false
	}
	if normalizedRequestPath != workspaceBasePath &&
		!strings.HasPrefix(normalizedRequestPath, workspaceBasePath+"/") {
		return "", false
	}
	assetPath := strings.TrimPrefix(normalizedRequestPath, workspaceBasePath)
	assetPath = strings.Trim(assetPath, "/")
	if assetPath == "" {
		return "index.html", true
	}
	return assetPath, true
}

// parsePluginAssetRequestPath splits one public `/x-assets/...` request
// path into plugin identity, version, and relative asset path parts.
func parsePluginAssetRequestPath(path string) (
	pluginID string,
	version string,
	assetPath string,
	ok bool,
) {
	normalizedPath := strings.Trim(strings.TrimSpace(path), "/")
	if normalizedPath == "" {
		return "", "", "", false
	}

	pathParts := strings.Split(normalizedPath, "/")
	if len(pathParts) < 3 || pathParts[0] != pluginhost.HostedAssetPathSegment {
		return "", "", "", false
	}
	if strings.TrimSpace(pathParts[1]) == "" || strings.TrimSpace(pathParts[2]) == "" {
		return "", "", "", false
	}

	pluginID = pathParts[1]
	version = pathParts[2]
	if len(pathParts) == 3 {
		return pluginID, version, "", true
	}
	return pluginID, version, strings.Join(pathParts[3:], "/"), true
}
