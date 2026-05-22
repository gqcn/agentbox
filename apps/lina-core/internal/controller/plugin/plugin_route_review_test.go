// This file tests dynamic-route review projections used by plugin management
// install and enable dialogs.

package plugin

import (
	"net/http"
	"testing"

	"lina-core/pkg/pluginbridge"
)

// TestBuildPluginRouteReviewItemsBuildsPublicRouteMetadata verifies current
// release route contracts are projected with host-visible public paths and
// review-friendly access metadata.
func TestBuildPluginRouteReviewItemsBuildsPublicRouteMetadata(t *testing.T) {
	items := buildPluginRouteReviewItems(
		"plugin-dev-route-review",
		[]*pluginbridge.RouteContract{
			{
				Method:      http.MethodGet,
				Path:        "/api/v1/review-summary",
				Access:      pluginbridge.AccessLogin,
				Permission:  "plugin-dev-route-review:review:query",
				Summary:     "查询评审摘要",
				Description: "返回插件当前评审摘要。",
			},
			nil,
			{
				Method:  http.MethodPost,
				Path:    "/api/v1/public-ping",
				Access:  pluginbridge.AccessPublic,
				Summary: "公开探活",
			},
		},
	)

	if len(items) != 2 {
		t.Fatalf("expected 2 projected route items, got %d", len(items))
	}

	if items[0].Method != http.MethodGet {
		t.Fatalf("expected first route method GET, got %s", items[0].Method)
	}
	if items[0].PublicPath != "/x/plugin-dev-route-review/api/v1/review-summary" {
		t.Fatalf("unexpected first route public path: %s", items[0].PublicPath)
	}
	if items[0].Access != pluginbridge.AccessLogin {
		t.Fatalf("expected first route access login, got %s", items[0].Access)
	}
	if items[0].Permission != "plugin-dev-route-review:review:query" {
		t.Fatalf("unexpected first route permission: %s", items[0].Permission)
	}
	if items[0].Summary != "查询评审摘要" {
		t.Fatalf("unexpected first route summary: %s", items[0].Summary)
	}
	if items[0].Description != "返回插件当前评审摘要。" {
		t.Fatalf("unexpected first route description: %s", items[0].Description)
	}

	if items[1].Method != http.MethodPost {
		t.Fatalf("expected second route method POST, got %s", items[1].Method)
	}
	if items[1].PublicPath != "/x/plugin-dev-route-review/api/v1/public-ping" {
		t.Fatalf("unexpected second route public path: %s", items[1].PublicPath)
	}
	if items[1].Access != pluginbridge.AccessPublic {
		t.Fatalf("expected second route access public, got %s", items[1].Access)
	}
	if items[1].Permission != "" {
		t.Fatalf("expected public route permission to stay empty, got %s", items[1].Permission)
	}
}

// TestBuildPluginRouteReviewItemsPreservesPluginOwnedPathContent verifies
// route review public paths only force `/x/{pluginId}` and preserve plugin-local
// path content such as `/api/v2`, `/interface/m1`, and `/graphql`.
func TestBuildPluginRouteReviewItemsPreservesPluginOwnedPathContent(t *testing.T) {
	items := buildPluginRouteReviewItems(
		"plugin-dev-route-review",
		[]*pluginbridge.RouteContract{
			{
				Method: http.MethodGet,
				Path:   "/api/v2/review-summary",
				Access: pluginbridge.AccessLogin,
			},
			{
				Method: http.MethodPost,
				Path:   "/interface/m1/review-summary",
				Access: pluginbridge.AccessLogin,
			},
			{
				Method: http.MethodPost,
				Path:   "/graphql",
				Access: pluginbridge.AccessPublic,
			},
		},
	)

	expected := []string{
		"/x/plugin-dev-route-review/api/v2/review-summary",
		"/x/plugin-dev-route-review/interface/m1/review-summary",
		"/x/plugin-dev-route-review/graphql",
	}
	if len(items) != len(expected) {
		t.Fatalf("expected %d projected route items, got %d", len(expected), len(items))
	}
	for index, expectedPath := range expected {
		if items[index].PublicPath != expectedPath {
			t.Fatalf("expected item %d public path %s, got %s", index, expectedPath, items[index].PublicPath)
		}
	}
}

// TestBuildPluginRouteReviewItemsSkipsBlankPluginID verifies blank plugin IDs
// do not produce invalid review items.
func TestBuildPluginRouteReviewItemsSkipsBlankPluginID(t *testing.T) {
	items := buildPluginRouteReviewItems(
		" ",
		[]*pluginbridge.RouteContract{
			{
				Method: http.MethodGet,
				Path:   "/api/v1/review-summary",
			},
		},
	)

	if items != nil {
		t.Fatalf("expected nil items for blank plugin ID, got %#v", items)
	}
}
