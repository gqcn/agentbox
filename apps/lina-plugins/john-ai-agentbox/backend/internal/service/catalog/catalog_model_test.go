// This file verifies pure catalog helpers used by provider model
// synchronization. Database-backed paths are covered by package compilation and
// route/controller tests in this migration slice.

package catalog

import (
	"net/http"
	"testing"
)

// TestProviderModelsURLNormalizesBaseURLs verifies resource URL construction avoids duplicate /v1/models.
func TestProviderModelsURLNormalizesBaseURLs(t *testing.T) {
	cases := map[string]string{
		"https://api.openai.com":            "https://api.openai.com/v1/models",
		"https://api.openai.com/v1":         "https://api.openai.com/v1/models",
		"https://api.openai.com/v1/models":  "https://api.openai.com/v1/models",
		"https://gateway.example/anthropic": "https://gateway.example/anthropic/v1/models",
	}
	for input, expected := range cases {
		actual, err := providerModelsURL(input, ProtocolOpenAI)
		if err != nil {
			t.Fatalf("providerModelsURL(%q): %v", input, err)
		}
		if actual != expected {
			t.Fatalf("providerModelsURL(%q)=%q, want %q", input, actual, expected)
		}
	}
}

// TestProviderAPIKeyHeaders verifies protocol-specific auth headers.
func TestProviderAPIKeyHeaders(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://example.test/v1/models", nil)
	if err != nil {
		t.Fatal(err)
	}
	setProviderAPIKeyHeaders(req, ProtocolOpenAI, "sk-test")
	if req.Header.Get("Authorization") != "Bearer sk-test" || req.Header.Get("api-key") != "sk-test" {
		t.Fatalf("unexpected OpenAI headers: %#v", req.Header)
	}

	req, err = http.NewRequest(http.MethodGet, "https://example.test/v1/models", nil)
	if err != nil {
		t.Fatal(err)
	}
	setProviderAPIKeyHeaders(req, ProtocolAnthropic, "ak-test")
	if req.Header.Get("x-api-key") != "ak-test" || req.Header.Get("api-key") != "ak-test" {
		t.Fatalf("unexpected Anthropic headers: %#v", req.Header)
	}
}

// TestNormalizeProtocol rejects unsupported protocol strings.
func TestNormalizeProtocol(t *testing.T) {
	if normalizeProtocol(" OPENAI ") != ProtocolOpenAI {
		t.Fatal("expected openai protocol normalization")
	}
	if normalizeProtocol("gemini") != "" {
		t.Fatal("expected unsupported protocol to normalize to empty string")
	}
}
