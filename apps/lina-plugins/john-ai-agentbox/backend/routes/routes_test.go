// This file verifies AgentBox source-plugin route registration keeps API
// routes inside the plugin namespace and only binds exact browser entry points.

package routes

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/guid"

	"lina-core/pkg/plugin/pluginhost"
)

// TestRegisterBindsRoutesUnderPluginNamespace verifies AgentBox API routes are
// registered through the source-plugin API prefix.
func TestRegisterBindsRoutesUnderPluginNamespace(t *testing.T) {
	server := g.Server("john-ai-agentbox-routes-test")
	var rootGroup *ghttp.RouterGroup
	server.Group("/", func(group *ghttp.RouterGroup) {
		rootGroup = group
	})

	registrar := pluginhost.NewHTTPRegistrar(
		server,
		rootGroup,
		"john-ai-agentbox",
		nil,
		nil,
		nil,
	)
	if err := Register(context.Background(), registrar, testPortalAssets()); err != nil {
		t.Fatalf("expected route registration to succeed, got %v", err)
	}

	bindings := registrar.Routes().RouteBindings()
	expected := map[string]struct{}{
		"GET /":      {},
		"GET /login": {},
		"POST /x/john-ai-agentbox/api/v1/auth/sessions":                                                              {},
		"GET /x/john-ai-agentbox/api/v1/auth/session":                                                                {},
		"DELETE /x/john-ai-agentbox/api/v1/auth/session":                                                             {},
		"GET /x/john-ai-agentbox/api/v1/ai/capability-tiers":                                                         {},
		"PUT /x/john-ai-agentbox/api/v1/ai/capability-tiers/{code}":                                                  {},
		"POST /x/john-ai-agentbox/api/v1/ai/capability-tiers/{code}/test":                                            {},
		"GET /x/john-ai-agentbox/api/v1/ai/invocations":                                                              {},
		"GET /x/john-ai-agentbox/api/v1/agents":                                                                      {},
		"POST /x/john-ai-agentbox/api/v1/agents":                                                                     {},
		"GET /x/john-ai-agentbox/api/v1/agents/{id}":                                                                 {},
		"PUT /x/john-ai-agentbox/api/v1/agents/{id}":                                                                 {},
		"PUT /x/john-ai-agentbox/api/v1/agents/{id}/image":                                                           {},
		"POST /x/john-ai-agentbox/api/v1/agents/{id}/start":                                                          {},
		"POST /x/john-ai-agentbox/api/v1/agents/{id}/stop":                                                           {},
		"GET /x/john-ai-agentbox/api/v1/agents/{id}/logs":                                                            {},
		"DELETE /x/john-ai-agentbox/api/v1/agents/{id}":                                                              {},
		"GET /x/john-ai-agentbox/api/v1/agents/{id}/services":                                                        {},
		"GET /x/john-ai-agentbox/api/v1/agents/{id}/services/{serviceId}":                                            {},
		"GET /x/john-ai-agentbox/api/v1/agents/{id}/service-bridges":                                                 {},
		"POST /x/john-ai-agentbox/api/v1/agents/{id}/service-bridges":                                                {},
		"DELETE /x/john-ai-agentbox/api/v1/agents/{id}/service-bridges/{bridgeId}":                                   {},
		"GET /x/john-ai-agentbox/api/v1/agents/{id}/chat/sessions":                                                   {},
		"POST /x/john-ai-agentbox/api/v1/agents/{id}/chat/sessions":                                                  {},
		"GET /x/john-ai-agentbox/api/v1/agents/{id}/chat/sessions/{sessionId}":                                       {},
		"PUT /x/john-ai-agentbox/api/v1/agents/{id}/chat/sessions/{sessionId}":                                       {},
		"DELETE /x/john-ai-agentbox/api/v1/agents/{id}/chat/sessions/{sessionId}":                                    {},
		"GET /x/john-ai-agentbox/api/v1/agents/{id}/chat/sessions/{sessionId}/messages":                              {},
		"POST /x/john-ai-agentbox/api/v1/agents/{id}/chat/sessions/{sessionId}/recover":                              {},
		"GET /x/john-ai-agentbox/api/v1/agents/{id}/chat/sessions/{sessionId}/interactions":                          {},
		"GET /x/john-ai-agentbox/api/v1/agents/{id}/chat/sessions/{sessionId}/interactions/{interactionId}":          {},
		"PUT /x/john-ai-agentbox/api/v1/agents/{id}/chat/sessions/{sessionId}/interactions/{interactionId}/response": {},
		"PUT /x/john-ai-agentbox/api/v1/agents/{id}/chat/sessions/{sessionId}/interactions/{interactionId}/status":   {},
		"GET /x/john-ai-agentbox/api/v1/agents/{id}/terminal/sessions":                                               {},
		"POST /x/john-ai-agentbox/api/v1/agents/{id}/terminal/sessions":                                              {},
		"GET /x/john-ai-agentbox/api/v1/agents/{id}/terminal/sessions/{terminalId}":                                  {},
		"DELETE /x/john-ai-agentbox/api/v1/agents/{id}/terminal/sessions/{terminalId}":                               {},
		"GET /x/john-ai-agentbox/api/v1/agents/{id}/workspace/paths":                                                 {},
		"GET /x/john-ai-agentbox/api/v1/agents/{id}/workspace/tree":                                                  {},
		"GET /x/john-ai-agentbox/api/v1/agents/{id}/workspace/file":                                                  {},
		"PUT /x/john-ai-agentbox/api/v1/agents/{id}/workspace/file":                                                  {},
		"POST /x/john-ai-agentbox/api/v1/agents/{id}/workspace/files":                                                {},
		"POST /x/john-ai-agentbox/api/v1/agents/{id}/workspace/directories":                                          {},
		"POST /x/john-ai-agentbox/api/v1/agents/{id}/workspace/upload":                                               {},
		"GET /x/john-ai-agentbox/api/v1/agents/{id}/workspace/download":                                              {},
		"GET /x/john-ai-agentbox/api/v1/agents/{id}/workspace/resources":                                             {},
		"GET /x/john-ai-agentbox/api/v1/agents/{id}/workspace/html-previews":                                         {},
		"GET /x/john-ai-agentbox/api/v1/agents/{id}/skills":                                                          {},
		"POST /x/john-ai-agentbox/api/v1/agents/{id}/skills/upload":                                                  {},
		"GET /x/john-ai-agentbox/api/v1/agents/{id}/git/status":                                                      {},
		"GET /x/john-ai-agentbox/api/v1/agents/{id}/git/file":                                                        {},
		"GET /x/john-ai-agentbox/api/v1/agents/{id}/git/diff":                                                        {},
		"POST /x/john-ai-agentbox/api/v1/agents/{id}/git/commit-message-suggestions":                                 {},
		"PUT /x/john-ai-agentbox/api/v1/agents/{id}/git/index":                                                       {},
		"DELETE /x/john-ai-agentbox/api/v1/agents/{id}/git/index":                                                    {},
		"DELETE /x/john-ai-agentbox/api/v1/agents/{id}/git/changes":                                                  {},
		"POST /x/john-ai-agentbox/api/v1/agents/{id}/git/commits":                                                    {},
		"GET /x/john-ai-agentbox/api/v1/health/docker":                                                               {},
		"GET /x/john-ai-agentbox/api/v1/containers":                                                                  {},
		"POST /x/john-ai-agentbox/api/v1/containers":                                                                 {},
		"GET /x/john-ai-agentbox/api/v1/containers/{id}":                                                             {},
		"POST /x/john-ai-agentbox/api/v1/containers/{id}/start":                                                      {},
		"POST /x/john-ai-agentbox/api/v1/containers/{id}/stop":                                                       {},
		"DELETE /x/john-ai-agentbox/api/v1/containers/{id}":                                                          {},
		"GET /x/john-ai-agentbox/api/v1/containers/{id}/logs":                                                        {},
		"ALL /x/john-ai-agentbox/api/v1/proxy/*":                                                                     {},
		"GET /x/john-ai-agentbox/api/v1/ws/agents/{id}/shell":                                                        {},
		"GET /x/john-ai-agentbox/api/v1/ws/agents/{id}/chat/sessions/{sessionId}":                                    {},
		"GET /x/john-ai-agentbox/api/v1/ws/agents/{id}/services/{serviceId}/tcp":                                     {},
		"GET /x/john-ai-agentbox/api/v1/providers":                                                                   {},
		"POST /x/john-ai-agentbox/api/v1/providers":                                                                  {},
		"GET /x/john-ai-agentbox/api/v1/providers/{id}":                                                              {},
		"PUT /x/john-ai-agentbox/api/v1/providers/{id}":                                                              {},
		"DELETE /x/john-ai-agentbox/api/v1/providers/{id}":                                                           {},
		"POST /x/john-ai-agentbox/api/v1/providers/{id}/models":                                                      {},
		"DELETE /x/john-ai-agentbox/api/v1/providers/{id}/models/{modelId}":                                          {},
		"POST /x/john-ai-agentbox/api/v1/providers/{id}/models/sync":                                                 {},
		"GET /x/john-ai-agentbox/api/v1/images":                                                                      {},
		"POST /x/john-ai-agentbox/api/v1/images":                                                                     {},
		"PUT /x/john-ai-agentbox/api/v1/images/{id}":                                                                 {},
		"DELETE /x/john-ai-agentbox/api/v1/images/{id}":                                                              {},
		"GET /x/john-ai-agentbox/api/v1/settings/{key}":                                                              {},
		"PUT /x/john-ai-agentbox/api/v1/settings/{key}":                                                              {},
		"GET /x/john-ai-agentbox/api/v1/prompt-templates":                                                            {},
		"GET /x/john-ai-agentbox/api/v1/prompt-templates/{code}":                                                     {},
		"PUT /x/john-ai-agentbox/api/v1/prompt-templates/{code}":                                                     {},
		"POST /x/john-ai-agentbox/api/v1/prompt-templates/{code}/restore":                                            {},
		"POST /x/john-ai-agentbox/api/v1/prompt-templates/{code}/previews":                                           {},
	}
	if len(bindings) != len(expected) {
		t.Fatalf("expected %d route bindings, got %d: %#v", len(expected), len(bindings), bindings)
	}
	rawRoutes := map[string]struct{}{
		"GET /":                                  {},
		"GET /login":                             {},
		"ALL /x/john-ai-agentbox/api/v1/proxy/*": {},
		"GET /x/john-ai-agentbox/api/v1/ws/agents/{id}/shell":                     {},
		"GET /x/john-ai-agentbox/api/v1/ws/agents/{id}/chat/sessions/{sessionId}": {},
		"GET /x/john-ai-agentbox/api/v1/ws/agents/{id}/services/{serviceId}/tcp":  {},
	}
	for _, binding := range bindings {
		if binding.PluginID != "john-ai-agentbox" {
			t.Fatalf("expected plugin id john-ai-agentbox, got %s", binding.PluginID)
		}
		key := binding.Method + " " + binding.Path
		if _, raw := rawRoutes[key]; raw {
			if binding.Documentable {
				t.Fatalf("expected raw route %s to be non-documentable", key)
			}
		} else if isAgentBoxPortalRouteBinding(binding.Path) {
			if binding.Documentable {
				t.Fatalf("expected browser entry route %s to be non-documentable", key)
			}
		} else if !binding.Documentable {
			t.Fatalf("expected %s to be documentable", key)
		}
		if _, ok := expected[key]; !ok {
			t.Fatalf("unexpected route binding %s", key)
		}
		delete(expected, key)
	}
	if len(expected) > 0 {
		t.Fatalf("missing route bindings: %#v", expected)
	}
}

// TestPortalRoutesServeIndexWithoutAssetRedirect verifies root browser entry
// points are the AgentBox workbench URLs instead of redirects to /x-assets.
func TestPortalRoutesServeIndexWithoutAssetRedirect(t *testing.T) {
	server := ghttp.GetServer("john-ai-agentbox-portal-test-" + guid.S())
	server.SetPort(0)
	server.SetDumpRouterMap(false)
	var rootGroup *ghttp.RouterGroup
	server.Group("/", func(group *ghttp.RouterGroup) {
		rootGroup = group
	})

	registrar := pluginhost.NewHTTPRegistrar(
		server,
		rootGroup,
		"john-ai-agentbox",
		nil,
		nil,
		nil,
	)
	if err := Register(context.Background(), registrar, testPortalAssets()); err != nil {
		t.Fatalf("expected route registration to succeed, got %v", err)
	}
	if err := server.Start(); err != nil {
		t.Fatalf("start portal route test server: %v", err)
	}
	t.Cleanup(func() {
		if err := server.Shutdown(); err != nil {
			t.Fatalf("shutdown portal route test server: %v", err)
		}
	})

	client := &http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	for _, route := range []string{"/", "/login"} {
		response, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d%s", server.GetListenedPort(), route))
		if err != nil {
			t.Fatalf("request %s: %v", route, err)
		}
		body, readErr := io.ReadAll(response.Body)
		closeErr := response.Body.Close()
		if readErr != nil {
			t.Fatalf("read %s response: %v", route, readErr)
		}
		if closeErr != nil {
			t.Fatalf("close %s response: %v", route, closeErr)
		}
		if response.StatusCode != http.StatusOK {
			t.Fatalf("expected %s to return 200, got %d location=%q", route, response.StatusCode, response.Header.Get("Location"))
		}
		if response.Header.Get("Location") != "" {
			t.Fatalf("expected %s not to redirect, got location %q", route, response.Header.Get("Location"))
		}
		if contentType := response.Header.Get("Content-Type"); !strings.HasPrefix(contentType, portalHTMLContentType) {
			t.Fatalf("expected %s content type %q, got %q", route, portalHTMLContentType, contentType)
		}
		if !strings.Contains(string(body), "agentbox-root") {
			t.Fatalf("expected %s to serve AgentBox index HTML, got %q", route, string(body))
		}
	}
}

func testPortalAssets() fstest.MapFS {
	return fstest.MapFS{
		portalIndexAssetPath: &fstest.MapFile{
			Data: []byte(`<!doctype html><html><body><div id="agentbox-root"></div></body></html>`),
		},
	}
}
