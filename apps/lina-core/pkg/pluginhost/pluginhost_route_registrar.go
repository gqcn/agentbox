// This file defines the public HTTP route registrar contract exposed to source
// plugins and the guarded host implementation that enforces plugin state.

package pluginhost

import (
	"context"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"
)

// sourceRouteCtxVarPluginID stores the source-plugin id that owns the matched route.
const sourceRouteCtxVarPluginID gctx.StrKey = "plugin_source_route_plugin_id"

// PluginEnabledChecker defines one host callback that reports whether a plugin
// is currently enabled in the supplied request or execution context.
type PluginEnabledChecker func(ctx context.Context, pluginID string) bool

// RouteMiddleware defines one plugin-usable HTTP middleware.
type RouteMiddleware = ghttp.HandlerFunc

// RouteMiddlewares exposes published host middlewares for plugin route composition.
type RouteMiddlewares interface {
	// NeverDoneCtx returns the host never-done context middleware.
	NeverDoneCtx() RouteMiddleware
	// HandlerResponse returns the host unified response middleware.
	HandlerResponse() RouteMiddleware
	// CORS returns the host CORS middleware.
	CORS() RouteMiddleware
	// RequestBodyLimit returns the host request-body limit middleware.
	RequestBodyLimit() RouteMiddleware
	// Ctx returns the host business-context injection middleware.
	Ctx() RouteMiddleware
	// Auth returns the host JWT authentication middleware.
	Auth() RouteMiddleware
	// Tenancy returns the host tenant-resolution middleware.
	Tenancy() RouteMiddleware
	// Permission returns the host declarative permission middleware.
	Permission() RouteMiddleware
}

// RouteRegistrar exposes plugin route group registration helpers for one plugin.
type RouteRegistrar interface {
	// APIPrefix returns the mandatory public API prefix for this plugin.
	APIPrefix() string
	// Err returns the first route registration error captured by the registrar.
	Err() error
	// Group registers one plugin route group bound to the dedicated plugin router root.
	Group(prefix string, register func(group RouteGroup))
	// Middlewares returns the published host middlewares available to plugins.
	Middlewares() RouteMiddlewares
	// RouteBindings returns the source-plugin route bindings captured by the host.
	RouteBindings() []SourceRouteBinding
}

// RouteGroup exposes one host-observable subset of GoFrame router-group
// registration methods for source plugins.
type RouteGroup interface {
	// Err returns the first route registration error captured by this group.
	Err() error
	// Group registers one nested route group.
	Group(prefix string, register func(group RouteGroup))
	// Middleware appends one or more middleware handlers to the current group.
	Middleware(handlers ...RouteMiddleware)
	// Bind registers one or more handlers or controller objects.
	Bind(handlerOrObject ...interface{})
	// ALL registers one handler for all HTTP methods.
	ALL(pattern string, object interface{}, params ...interface{})
	// GET registers one GET handler.
	GET(pattern string, object interface{}, params ...interface{})
	// PUT registers one PUT handler.
	PUT(pattern string, object interface{}, params ...interface{})
	// POST registers one POST handler.
	POST(pattern string, object interface{}, params ...interface{})
	// DELETE registers one DELETE handler.
	DELETE(pattern string, object interface{}, params ...interface{})
	// PATCH registers one PATCH handler.
	PATCH(pattern string, object interface{}, params ...interface{})
	// HEAD registers one HEAD handler.
	HEAD(pattern string, object interface{}, params ...interface{})
	// CONNECT registers one CONNECT handler.
	CONNECT(pattern string, object interface{}, params ...interface{})
	// OPTIONS registers one OPTIONS handler.
	OPTIONS(pattern string, object interface{}, params ...interface{})
	// TRACE registers one TRACE handler.
	TRACE(pattern string, object interface{}, params ...interface{})
}

// routeRegistrar is the host-owned RouteRegistrar implementation for one source
// plugin registration session.
type routeRegistrar struct {
	group          *ghttp.RouterGroup
	pluginID       string
	enabledChecker PluginEnabledChecker
	middlewares    RouteMiddlewares
	bindingsMu     sync.RWMutex
	bindings       []SourceRouteBinding
	errMu          sync.RWMutex
	err            error
}

// routeGroup adapts one GoFrame router group to the reduced plugin RouteGroup
// contract while preserving host-side route capture.
type routeGroup struct {
	group     *ghttp.RouterGroup
	pluginID  string
	prefix    string
	bindRoute func(bindings ...SourceRouteBinding)
	setError  func(error)
	getError  func() error
}

// routeMiddlewares stores the published host middleware directory that source
// plugins are allowed to reuse.
type routeMiddlewares struct {
	neverDoneCtx    RouteMiddleware
	handlerResponse RouteMiddleware
	cors            RouteMiddleware
	requestBody     RouteMiddleware
	ctx             RouteMiddleware
	auth            RouteMiddleware
	tenancy         RouteMiddleware
	permission      RouteMiddleware
}

// NewRouteMiddlewares creates and returns a new published host middleware directory for plugins.
func NewRouteMiddlewares(
	neverDoneCtx RouteMiddleware,
	handlerResponse RouteMiddleware,
	cors RouteMiddleware,
	requestBody RouteMiddleware,
	ctx RouteMiddleware,
	auth RouteMiddleware,
	tenancy RouteMiddleware,
	permission RouteMiddleware,
) RouteMiddlewares {
	return &routeMiddlewares{
		neverDoneCtx:    neverDoneCtx,
		handlerResponse: handlerResponse,
		cors:            cors,
		requestBody:     requestBody,
		ctx:             ctx,
		auth:            auth,
		tenancy:         tenancy,
		permission:      permission,
	}
}

// SourcePluginIDFromRequest returns the source-plugin id attached to the current
// request, or an empty string when the request is not handled by a source plugin.
func SourcePluginIDFromRequest(request *ghttp.Request) string {
	if request == nil {
		return ""
	}
	return strings.TrimSpace(request.GetCtxVar(sourceRouteCtxVarPluginID).String())
}

// NewRouteRegistrar creates and returns a new RouteRegistrar instance.
func NewRouteRegistrar(
	pluginGroup *ghttp.RouterGroup,
	pluginID string,
	enabledChecker PluginEnabledChecker,
	middlewares RouteMiddlewares,
) RouteRegistrar {
	return &routeRegistrar{
		group:          pluginGroup,
		pluginID:       pluginID,
		enabledChecker: enabledChecker,
		middlewares:    middlewares,
	}
}

// APIPrefix returns the mandatory public API prefix for this plugin.
func (r *routeRegistrar) APIPrefix() string {
	if r == nil {
		return ""
	}
	return pluginAPIPrefix(r.pluginID)
}

// Err returns the first route registration error captured by this registrar.
func (r *routeRegistrar) Err() error {
	if r == nil {
		return nil
	}
	r.errMu.RLock()
	defer r.errMu.RUnlock()
	return r.err
}

// Group registers one plugin route group bound to the dedicated plugin router root.
func (r *routeRegistrar) Group(prefix string, register func(group RouteGroup)) {
	if r == nil || r.group == nil || register == nil {
		return
	}

	normalizedPrefix := normalizeRoutePrefix(prefix)
	if err := validateSourceRoutePrefix(r.pluginID, normalizedPrefix); err != nil {
		r.setError(err)
		return
	}
	r.group.Group(normalizedPrefix, func(group *ghttp.RouterGroup) {
		group.Middleware(func(req *ghttp.Request) {
			if !r.allow(req) {
				return
			}
			req.Middleware.Next()
		})
		register(&routeGroup{
			group:     group,
			pluginID:  r.pluginID,
			prefix:    normalizedPrefix,
			bindRoute: r.appendBindings,
			setError:  r.setError,
			getError:  r.Err,
		})
	})
}

// Middlewares returns the published host middlewares available to plugins.
func (r *routeRegistrar) Middlewares() RouteMiddlewares {
	if r == nil {
		return nil
	}
	return r.middlewares
}

// RouteBindings returns the source-plugin route bindings captured by the host.
func (r *routeRegistrar) RouteBindings() []SourceRouteBinding {
	if r == nil {
		return nil
	}
	r.bindingsMu.RLock()
	defer r.bindingsMu.RUnlock()
	return CloneSourceRouteBindings(r.bindings)
}

// allow rejects plugin route traffic when the owning plugin is not currently
// enabled, matching the host-side plugin state governance used elsewhere.
func (r *routeRegistrar) allow(req *ghttp.Request) bool {
	if req == nil {
		return false
	}
	if r.enabledChecker != nil && !r.enabledChecker(req.Context(), r.pluginID) {
		req.Response.WriteStatus(404)
		req.ExitAll()
		return false
	}
	req.SetCtxVar(sourceRouteCtxVarPluginID, strings.TrimSpace(r.pluginID))
	return true
}

// NeverDoneCtx returns the published never-done context middleware.
func (m *routeMiddlewares) NeverDoneCtx() RouteMiddleware {
	if m == nil {
		return nil
	}
	return m.neverDoneCtx
}

// HandlerResponse returns the published unified response middleware.
func (m *routeMiddlewares) HandlerResponse() RouteMiddleware {
	if m == nil {
		return nil
	}
	return m.handlerResponse
}

// CORS returns the published CORS middleware.
func (m *routeMiddlewares) CORS() RouteMiddleware {
	if m == nil {
		return nil
	}
	return m.cors
}

// RequestBodyLimit returns the published request-body limit middleware.
func (m *routeMiddlewares) RequestBodyLimit() RouteMiddleware {
	if m == nil {
		return nil
	}
	return m.requestBody
}

// Ctx returns the published business-context injection middleware.
func (m *routeMiddlewares) Ctx() RouteMiddleware {
	if m == nil {
		return nil
	}
	return m.ctx
}

// Auth returns the published authentication middleware.
func (m *routeMiddlewares) Auth() RouteMiddleware {
	if m == nil {
		return nil
	}
	return m.auth
}

// Tenancy returns the published tenant-resolution middleware.
func (m *routeMiddlewares) Tenancy() RouteMiddleware {
	if m == nil {
		return nil
	}
	return m.tenancy
}

// Permission returns the published declarative permission middleware.
func (m *routeMiddlewares) Permission() RouteMiddleware {
	if m == nil {
		return nil
	}
	return m.permission
}

// normalizeRoutePrefix canonicalizes plugin-owned route group prefixes so
// callers can register `api/v1`, `/api/v1/`, or `/` interchangeably.
func normalizeRoutePrefix(prefix string) string {
	trimmed := strings.TrimSpace(prefix)
	if trimmed == "" || trimmed == "/" {
		return "/"
	}
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	return strings.TrimRight(trimmed, "/")
}

// pluginAPIPrefix returns the mandatory plugin-owned namespace for one plugin ID.
func pluginAPIPrefix(pluginID string) string {
	return PluginAPINamespacePrefix + "/" + strings.Trim(strings.TrimSpace(pluginID), "/")
}

// validateSourceRoutePrefix rejects source-plugin registrations that try to use
// `/x` for anything outside the owning plugin namespace.
func validateSourceRoutePrefix(pluginID string, prefix string) error {
	normalizedPrefix := normalizeRoutePrefix(prefix)
	return validateSourceRoutePath(pluginID, normalizedPrefix)
}

// validateSourceRoutePath rejects final source-plugin route paths that try to
// use `/x` for anything outside the owning plugin namespace.
func validateSourceRoutePath(pluginID string, routePath string) error {
	if routePathHasTraversal(routePath) {
		return gerror.Newf("source plugin %s cannot register route with traversal segments: %s", strings.TrimSpace(pluginID), routePath)
	}
	normalizedPrefix := normalizeRoutePrefix(routePath)
	if normalizedPrefix == "/" {
		return nil
	}
	if normalizedPrefix != PluginAPINamespacePrefix && !strings.HasPrefix(normalizedPrefix, PluginAPINamespacePrefix+"/") {
		return nil
	}
	apiPrefix := pluginAPIPrefix(pluginID)
	if normalizedPrefix == apiPrefix || strings.HasPrefix(normalizedPrefix, apiPrefix+"/") {
		return nil
	}
	return gerror.Newf(
		"source plugin %s cannot register route outside its /x namespace: %s; use %s for plugin routes",
		strings.TrimSpace(pluginID),
		normalizedPrefix,
		apiPrefix,
	)
}

// routePathHasTraversal reports whether a plugin-supplied route contains dot
// segments that could change the effective reserved namespace after cleaning.
func routePathHasTraversal(routePath string) bool {
	parts := strings.Split(strings.ReplaceAll(strings.TrimSpace(routePath), "\\", "/"), "/")
	for _, part := range parts {
		if part == ".." || part == "." {
			return true
		}
	}
	return false
}

// appendBindings appends captured plugin route bindings to the registrar-local
// snapshot in a thread-safe way.
func (r *routeRegistrar) appendBindings(bindings ...SourceRouteBinding) {
	if r == nil || len(bindings) == 0 {
		return
	}
	r.bindingsMu.Lock()
	defer r.bindingsMu.Unlock()
	r.bindings = append(r.bindings, bindings...)
}

// setError records the first route registration error for the registrar.
func (r *routeRegistrar) setError(err error) {
	if r == nil || err == nil {
		return
	}
	r.errMu.Lock()
	defer r.errMu.Unlock()
	if r.err == nil {
		r.err = err
	}
}

// Err returns the first route registration error captured by this group.
func (g *routeGroup) Err() error {
	if g == nil || g.getError == nil {
		return nil
	}
	return g.getError()
}

// Group registers one nested route group.
func (g *routeGroup) Group(prefix string, register func(group RouteGroup)) {
	if g == nil || g.group == nil || register == nil {
		return
	}
	normalizedPrefix := normalizeRoutePrefix(prefix)
	joinedPrefix := joinRoutePatterns(g.prefix, normalizedPrefix)
	if err := validateSourceRoutePath(g.pluginID, joinedPrefix); err != nil {
		g.setRegistrationError(err)
		return
	}
	g.group.Group(normalizedPrefix, func(group *ghttp.RouterGroup) {
		register(&routeGroup{
			group:     group,
			pluginID:  g.pluginID,
			prefix:    joinedPrefix,
			bindRoute: g.bindRoute,
			setError:  g.setError,
			getError:  g.getError,
		})
	})
}

// Middleware appends one or more middleware handlers to the current group.
func (g *routeGroup) Middleware(handlers ...RouteMiddleware) {
	if g == nil || g.group == nil {
		return
	}
	g.group.Middleware(handlers...)
}

// Bind registers one or more handlers or controller objects.
func (g *routeGroup) Bind(handlerOrObject ...interface{}) {
	if g == nil || g.group == nil {
		return
	}
	for _, item := range handlerOrObject {
		bindings := captureRouteBindings(g.pluginID, g.prefix, "/", routeMethodAll, item)
		if err := validateSourceRouteBindings(g.pluginID, bindings); err != nil {
			g.setRegistrationError(err)
			return
		}
		if g.bindRoute != nil && len(bindings) > 0 {
			g.bindRoute(bindings...)
		}
	}
	g.group.Bind(handlerOrObject...)
}

// ALL registers one handler for all HTTP methods.
func (g *routeGroup) ALL(pattern string, object interface{}, params ...interface{}) {
	g.bindMethodRoute(pattern, routeMethodAll, object, params, func() {
		g.group.ALL(pattern, object, params...)
	})
}

// GET registers one GET handler.
func (g *routeGroup) GET(pattern string, object interface{}, params ...interface{}) {
	g.bindMethodRoute(pattern, "GET", object, params, func() {
		g.group.GET(pattern, object, params...)
	})
}

// PUT registers one PUT handler.
func (g *routeGroup) PUT(pattern string, object interface{}, params ...interface{}) {
	g.bindMethodRoute(pattern, "PUT", object, params, func() {
		g.group.PUT(pattern, object, params...)
	})
}

// POST registers one POST handler.
func (g *routeGroup) POST(pattern string, object interface{}, params ...interface{}) {
	g.bindMethodRoute(pattern, "POST", object, params, func() {
		g.group.POST(pattern, object, params...)
	})
}

// DELETE registers one DELETE handler.
func (g *routeGroup) DELETE(pattern string, object interface{}, params ...interface{}) {
	g.bindMethodRoute(pattern, "DELETE", object, params, func() {
		g.group.DELETE(pattern, object, params...)
	})
}

// PATCH registers one PATCH handler.
func (g *routeGroup) PATCH(pattern string, object interface{}, params ...interface{}) {
	g.bindMethodRoute(pattern, "PATCH", object, params, func() {
		g.group.PATCH(pattern, object, params...)
	})
}

// HEAD registers one HEAD handler.
func (g *routeGroup) HEAD(pattern string, object interface{}, params ...interface{}) {
	g.bindMethodRoute(pattern, "HEAD", object, params, func() {
		g.group.HEAD(pattern, object, params...)
	})
}

// CONNECT registers one CONNECT handler.
func (g *routeGroup) CONNECT(pattern string, object interface{}, params ...interface{}) {
	g.bindMethodRoute(pattern, "CONNECT", object, params, func() {
		g.group.CONNECT(pattern, object, params...)
	})
}

// OPTIONS registers one OPTIONS handler.
func (g *routeGroup) OPTIONS(pattern string, object interface{}, params ...interface{}) {
	g.bindMethodRoute(pattern, "OPTIONS", object, params, func() {
		g.group.OPTIONS(pattern, object, params...)
	})
}

// TRACE registers one TRACE handler.
func (g *routeGroup) TRACE(pattern string, object interface{}, params ...interface{}) {
	g.bindMethodRoute(pattern, "TRACE", object, params, func() {
		g.group.TRACE(pattern, object, params...)
	})
}

// bindMethodRoute captures one explicit method route before delegating to the
// underlying GoFrame router group.
func (g *routeGroup) bindMethodRoute(
	pattern string,
	method string,
	object interface{},
	_ []interface{},
	bind func(),
) {
	if g == nil || g.group == nil || bind == nil {
		return
	}
	bindings := captureRouteBindings(g.pluginID, g.prefix, pattern, method, object)
	if len(bindings) == 0 {
		finalPath := joinRoutePatterns(g.prefix, pattern)
		if err := validateSourceRoutePath(g.pluginID, finalPath); err != nil {
			g.setRegistrationError(err)
			return
		}
	}
	if err := validateSourceRouteBindings(g.pluginID, bindings); err != nil {
		g.setRegistrationError(err)
		return
	}
	if g.bindRoute != nil {
		g.bindRoute(bindings...)
	}
	bind()
}

// setRegistrationError records one invalid route registration on the owning registrar.
func (g *routeGroup) setRegistrationError(err error) {
	if g == nil || err == nil || g.setError == nil {
		return
	}
	g.setError(err)
}

// validateSourceRouteBindings checks documentable DTO routes after their
// GoFrame metadata has resolved to final public paths.
func validateSourceRouteBindings(pluginID string, bindings []SourceRouteBinding) error {
	for _, binding := range bindings {
		if err := validateSourceRoutePath(pluginID, binding.Path); err != nil {
			return err
		}
	}
	return nil
}
