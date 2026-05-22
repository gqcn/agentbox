// This file contains unit tests for the route registrar helpers exposed by the
// pluginhost package to source plugins and host bootstrap code.

package pluginhost

import (
	"context"
	"fmt"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// testPluginPingReq defines one strict-route DTO used by route binding tests.
type testPluginPingReq struct {
	g.Meta `path:"/plugins/test-plugin/ping" method:"get"`
}

// testPluginPingRes is the response DTO for the strict-route ping handler.
type testPluginPingRes struct{}

// testPluginIllegalXReq defines one invalid DTO route outside the owned /x namespace.
type testPluginIllegalXReq struct {
	g.Meta `path:"/x/other-plugin/health" method:"get"`
}

// testPluginIllegalXRes is the response DTO for an invalid /x route.
type testPluginIllegalXRes struct{}

// testPluginPingHandler is the strict-route handler used to verify route capture.
func testPluginPingHandler(ctx context.Context, req *testPluginPingReq) (*testPluginPingRes, error) {
	return &testPluginPingRes{}, nil
}

// testPluginIllegalXHandler is the strict-route handler used to verify /x namespace validation.
func testPluginIllegalXHandler(ctx context.Context, req *testPluginIllegalXReq) (*testPluginIllegalXRes, error) {
	return &testPluginIllegalXRes{}, nil
}

// TestNewRouteRegistrarExposeRootGroupAndPublishedMiddlewares verifies the
// registrar exposes the root group and the published middleware directory.
func TestNewRouteRegistrarExposeRootGroupAndPublishedMiddlewares(t *testing.T) {
	noop := func(r *ghttp.Request) {}
	middlewares := NewRouteMiddlewares(noop, noop, noop, noop, noop, noop, noop, noop)
	server := g.Server("pluginhost-route-registrar-test")

	var rootGroup *ghttp.RouterGroup
	server.Group("/", func(group *ghttp.RouterGroup) {
		rootGroup = group
	})

	registrar := NewRouteRegistrar(rootGroup, "plugin-demo", nil, middlewares)
	typed, ok := registrar.(*routeRegistrar)
	if !ok {
		t.Fatalf("expected concrete route registrar type")
	}
	if typed.group == nil {
		t.Fatalf("expected root plugin route group to be initialized")
	}
	if registrar.Middlewares() == nil {
		t.Fatalf("expected published middleware directory to be available")
	}
	if registrar.APIPrefix() != "/x/plugin-demo" {
		t.Fatalf("expected plugin API prefix /x/plugin-demo, got %q", registrar.APIPrefix())
	}

	called := false
	registrar.Group(registrar.APIPrefix()+"/api/v1", func(group RouteGroup) {
		called = true
		if group == nil {
			t.Fatalf("expected callback group to be initialized")
		}
		group.Middleware(
			middlewares.RequestBodyLimit(),
			middlewares.Ctx(),
			middlewares.Auth(),
			middlewares.Permission(),
		)
	})
	if !called {
		t.Fatalf("expected group callback to be invoked during route registration")
	}
}

// TestRouteRegistrarCaptureSourceRouteBindings verifies the host captures
// plugin-owned route bindings while preserving nested group path composition.
func TestRouteRegistrarCaptureSourceRouteBindings(t *testing.T) {
	server := g.Server("pluginhost-route-registrar-binding-test")

	var rootGroup *ghttp.RouterGroup
	server.Group("/", func(group *ghttp.RouterGroup) {
		rootGroup = group
	})

	registrar := NewRouteRegistrar(rootGroup, "plugin-demo", nil, nil)
	registrar.Group(registrar.APIPrefix()+"/api/v1", func(group RouteGroup) {
		group.Group("/plugins", func(group RouteGroup) {
			group.Bind(testPluginPingHandler)
			group.GET("/raw", func(r *ghttp.Request) {})
		})
	})

	bindings := registrar.RouteBindings()
	if len(bindings) != 2 {
		t.Fatalf("expected 2 route bindings, got %d", len(bindings))
	}
	if bindings[0].PluginID != "plugin-demo" {
		t.Fatalf("expected plugin id plugin-demo, got %s", bindings[0].PluginID)
	}
	if bindings[0].Method != "GET" {
		t.Fatalf("expected GET binding, got %s", bindings[0].Method)
	}
	if bindings[0].Path != "/x/plugin-demo/api/v1/plugins/plugins/test-plugin/ping" {
		t.Fatalf("expected strict route path to include nested prefix, got %s", bindings[0].Path)
	}
	if !bindings[0].Documentable {
		t.Fatalf("expected strict DTO handler to be documentable")
	}
	if bindings[1].Path != "/x/plugin-demo/api/v1/plugins/raw" {
		t.Fatalf("expected raw handler path /x/plugin-demo/api/v1/plugins/raw, got %s", bindings[1].Path)
	}
	if bindings[1].Documentable {
		t.Fatalf("expected raw handler to be non-documentable")
	}
}

// TestRouteRegistrarRejectsNonAPIRoutesUnderX verifies `/x` remains reserved
// for plugin API routes owned by the current plugin.
func TestRouteRegistrarRejectsNonAPIRoutesUnderX(t *testing.T) {
	server := g.Server("pluginhost-route-x-guard-test")

	var rootGroup *ghttp.RouterGroup
	server.Group("/", func(group *ghttp.RouterGroup) {
		rootGroup = group
	})

	registrar := NewRouteRegistrar(rootGroup, "plugin-demo", nil, nil)
	registrar.Group(registrar.APIPrefix()+"/api/v1", func(group RouteGroup) {
		group.GET("/ping", func(request *ghttp.Request) {})
	})

	tests := []struct {
		name   string
		prefix string
	}{
		{name: "x root", prefix: "/x"},
		{name: "other plugin api", prefix: "/x/other-plugin/api/v1"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			registrar.Group(testCase.prefix, func(group RouteGroup) {
				group.GET("/ping", func(request *ghttp.Request) {})
			})
			if registrar.Err() == nil {
				t.Fatalf("expected prefix %q to be rejected", testCase.prefix)
			}
		})
	}
}

// TestRouteRegistrarAllowsPluginOwnedPathsUnderX verifies `/x/{pluginId}` is
// the only mandatory source-plugin API prefix and plugin-local route content is
// not forced to use `/api/v1`.
func TestRouteRegistrarAllowsPluginOwnedPathsUnderX(t *testing.T) {
	server := g.Server("pluginhost-route-x-owned-path-test")

	var rootGroup *ghttp.RouterGroup
	server.Group("/", func(group *ghttp.RouterGroup) {
		rootGroup = group
	})

	tests := []struct {
		name         string
		groupPrefix  string
		routePattern string
		expectedPath string
	}{
		{
			name:         "api v2",
			groupPrefix:  "/x/plugin-demo/api/v2",
			routePattern: "/items",
			expectedPath: "/x/plugin-demo/api/v2/items",
		},
		{
			name:         "interface path",
			groupPrefix:  "/x/plugin-demo/interface/m1",
			routePattern: "/items",
			expectedPath: "/x/plugin-demo/interface/m1/items",
		},
		{
			name:         "graphql",
			groupPrefix:  "/x/plugin-demo",
			routePattern: "/graphql",
			expectedPath: "/x/plugin-demo/graphql",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			registrar := NewRouteRegistrar(rootGroup, "plugin-demo", nil, nil)
			registrar.Group(testCase.groupPrefix, func(group RouteGroup) {
				group.GET(testCase.routePattern, func(request *ghttp.Request) {})
			})
			if err := registrar.Err(); err != nil {
				t.Fatalf("expected plugin-owned /x route to be accepted, got error: %v", err)
			}
			bindings := registrar.RouteBindings()
			if len(bindings) != 1 {
				t.Fatalf("expected 1 route binding, got %d", len(bindings))
			}
			if bindings[0].Path != testCase.expectedPath {
				t.Fatalf("expected route path %s, got %s", testCase.expectedPath, bindings[0].Path)
			}
		})
	}
}

// TestRouteRegistrarRejectsNestedRoutesOutsideOwnedX verifies nested groups and
// DTO metadata cannot bypass the reserved /x plugin namespace guard.
func TestRouteRegistrarRejectsNestedRoutesOutsideOwnedX(t *testing.T) {
	server := g.Server("pluginhost-route-x-nested-guard-test")

	var rootGroup *ghttp.RouterGroup
	server.Group("/", func(group *ghttp.RouterGroup) {
		rootGroup = group
	})

	tests := []struct {
		name     string
		register func(registrar RouteRegistrar)
	}{
		{
			name: "nested group",
			register: func(registrar RouteRegistrar) {
				registrar.Group("/x/plugin-demo/api/v1", func(group RouteGroup) {
					group.Group("/../../x/other-plugin/assets", func(group RouteGroup) {
						group.GET("/logo", func(request *ghttp.Request) {})
					})
				})
			},
		},
		{
			name: "dto meta path",
			register: func(registrar RouteRegistrar) {
				registrar.Group("/", func(group RouteGroup) {
					group.Bind(testPluginIllegalXHandler)
				})
			},
		},
		{
			name: "method path",
			register: func(registrar RouteRegistrar) {
				registrar.Group("/", func(group RouteGroup) {
					group.GET("/x/other-plugin/health", func(request *ghttp.Request) {})
				})
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			registrar := NewRouteRegistrar(rootGroup, "plugin-demo", nil, nil)
			testCase.register(registrar)
			if registrar.Err() == nil {
				t.Fatalf("expected route registration to be rejected")
			}
		})
	}
}

// TestNormalizeRoutePrefix verifies plugin route prefixes are normalized before
// the host binds them to a router group.
func TestNormalizeRoutePrefix(t *testing.T) {
	if got := normalizeRoutePrefix("api/v1/"); got != "/api/v1" {
		t.Fatalf("expected normalized prefix /api/v1, got %s", got)
	}
	if got := normalizeRoutePrefix(""); got != "/" {
		t.Fatalf("expected empty prefix to normalize to root, got %s", got)
	}
}

// TestRouteRegistrarEnabledCheckerReceivesRequestContext verifies route guards
// pass the active request context into the plugin-state checker.
func TestRouteRegistrarEnabledCheckerReceivesRequestContext(t *testing.T) {
	server := g.Server("pluginhost-route-context-checker-test")
	server.SetDumpRouterMap(false)
	server.SetPort(0)

	type contextKey struct{}
	ctxKey := contextKey{}
	var rootGroup *ghttp.RouterGroup
	server.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(func(request *ghttp.Request) {
			request.SetCtx(context.WithValue(request.Context(), ctxKey, "tenant-visible"))
			request.Middleware.Next()
		})
		rootGroup = group
	})

	checkerSawContext := false
	registrar := NewRouteRegistrar(rootGroup, "plugin-demo", func(ctx context.Context, pluginID string) bool {
		checkerSawContext = pluginID == "plugin-demo" && ctx.Value(ctxKey) == "tenant-visible"
		return true
	}, nil)
	registrar.Group("/plugins/plugin-demo", func(group RouteGroup) {
		group.GET("/ping", func(request *ghttp.Request) {
			request.Response.Write("ok")
		})
	})

	server.Start()
	defer server.Shutdown()

	client := g.Client()
	client.SetPrefix(fmt.Sprintf("http://127.0.0.1:%d", server.GetListenedPort()))
	if body := client.GetContent(context.Background(), "/plugins/plugin-demo/ping"); body != "ok" {
		t.Fatalf("expected route response ok, got %q", body)
	}
	if !checkerSawContext {
		t.Fatal("expected checker to receive active request context")
	}
}
