// This file covers OpenAPI helper projection logic for dynamic plugin routes.

package openapi

import (
	"net/http"
	"testing"

	"lina-core/pkg/pluginbridge"
)

// TestBuildRouteOpenAPIOperationUsesBridgeState verifies that projected
// responses follow the runtime bridge execution flag.
func TestBuildRouteOpenAPIOperationUsesBridgeState(t *testing.T) {
	operation := BuildRouteOpenAPIOperation("linapro-demo-dynamic", &pluginbridge.RouteContract{
		Path:    "/api/v1/review-summary",
		Method:  http.MethodGet,
		Access:  pluginbridge.AccessLogin,
		Summary: "Review Summary",
	}, &pluginbridge.BridgeSpec{
		RouteExecution: true,
	})
	if operation == nil || operation.Responses["200"].Value == nil {
		t.Fatalf("expected executable bridge operation to expose 200 response, got %#v", operation)
	}
	if operation.Responses["500"].Value == nil {
		t.Fatalf("expected executable bridge operation to expose 500 response, got %#v", operation)
	}
	if operation.Responses["501"].Value != nil {
		t.Fatalf("expected executable bridge operation to hide 501 placeholder response, got %#v", operation)
	}
	if operation.Security == nil {
		t.Fatal("expected login route to project bearer security")
	}
	if operation.OperationID != "" {
		t.Fatalf("expected dynamic OpenAPI operationId to stay empty, got %s", operation.OperationID)
	}
	if len(operation.XExtensions) != 0 {
		t.Fatalf("expected dynamic OpenAPI operation to omit i18n extensions, got %#v", operation.XExtensions)
	}

	placeholder := BuildRouteOpenAPIOperation("linapro-demo-dynamic", &pluginbridge.RouteContract{
		Path:   "/api/v1/placeholder",
		Method: http.MethodGet,
		Access: pluginbridge.AccessPublic,
	}, &pluginbridge.BridgeSpec{
		RouteExecution: false,
	})
	if placeholder == nil || placeholder.Responses["501"].Value == nil {
		t.Fatalf("expected placeholder bridge operation to expose 501 response, got %#v", placeholder)
	}
	if placeholder.Responses["200"].Value != nil {
		t.Fatalf("expected placeholder bridge operation to omit 200 response, got %#v", placeholder)
	}
}

// TestBuildRoutePublicPathBuildsFixedPublicPath verifies that public route
// projection always uses the canonical dynamic plugin prefix.
func TestBuildRoutePublicPathBuildsFixedPublicPath(t *testing.T) {
	actual := BuildRoutePublicPath("plugin-openapi-projection", "/api/v1/review-summary")
	if actual != "/x/plugin-openapi-projection/api/v1/review-summary" {
		t.Fatalf("expected fixed public path projection, got %s", actual)
	}
}

// TestBuildRoutePublicPathPreservesPluginOwnedPathContent verifies `/api/v1`
// is only a plugin-local naming convention and not a forced public-path segment.
func TestBuildRoutePublicPathPreservesPluginOwnedPathContent(t *testing.T) {
	tests := []struct {
		name      string
		routePath string
		expected  string
	}{
		{
			name:      "api v2",
			routePath: "/api/v2/review-summary",
			expected:  "/x/plugin-openapi-projection/api/v2/review-summary",
		},
		{
			name:      "interface",
			routePath: "/interface/m1/review-summary",
			expected:  "/x/plugin-openapi-projection/interface/m1/review-summary",
		},
		{
			name:      "graphql",
			routePath: "/graphql",
			expected:  "/x/plugin-openapi-projection/graphql",
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			actual := BuildRoutePublicPath("plugin-openapi-projection", testCase.routePath)
			if actual != testCase.expected {
				t.Fatalf("expected public path %s, got %s", testCase.expected, actual)
			}
		})
	}
}
