// This file verifies provider test adapters with fake HTTP servers so no real
// provider network calls are required.

package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"john-ai-agentbox/backend/internal/service/catalog"
)

// TestGenerateOpenAIRoutesRequest verifies OpenAI-compatible test requests.
func TestGenerateOpenAIRoutesRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != openAIChatCompletionsPath {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer openai-secret" {
			t.Fatalf("Authorization = %q", got)
		}
		var payload openAIChatRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload.Model != "gpt-test" || len(payload.Messages) != 2 || payload.Messages[1].Content != capabilityTestPrompt {
			t.Fatalf("payload = %#v", payload)
		}
		if _, err := w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}]}`)); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	service := &serviceImpl{httpClient: server.Client()}
	out, err := service.generateOpenAI(context.Background(), ResolvedBinding{
		Binding: CapabilityBindingInfo{
			ModelName: "gpt-test",
			Protocol:  catalog.ProtocolOpenAI,
		},
		Provider: providerRecord{
			APIKey:        "openai-secret",
			OpenAIBaseURL: server.URL,
		},
	}, capabilityTestPrompt, capabilityTestMaxOutputTokens)
	if err != nil {
		t.Fatal(err)
	}
	if out.Text != "ok" {
		t.Fatalf("text = %q", out.Text)
	}
}

// TestGenerateAnthropicRoutesRequest verifies Anthropic-compatible test requests.
func TestGenerateAnthropicRoutesRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != anthropicMessagesPath {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if got := r.Header.Get("x-api-key"); got != "anthropic-secret" {
			t.Fatalf("x-api-key = %q", got)
		}
		if got := r.Header.Get("anthropic-version"); got != anthropicVersion {
			t.Fatalf("anthropic-version = %q", got)
		}
		var payload anthropicMessagesRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload.Model != "claude-test" || len(payload.Messages) != 1 || payload.Messages[0].Content != capabilityTestPrompt {
			t.Fatalf("payload = %#v", payload)
		}
		if _, err := w.Write([]byte(`{"content":[{"type":"text","text":"ok"}]}`)); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	service := &serviceImpl{httpClient: server.Client()}
	out, err := service.generateAnthropic(context.Background(), ResolvedBinding{
		Binding: CapabilityBindingInfo{
			ModelName: "claude-test",
			Protocol:  catalog.ProtocolAnthropic,
		},
		Provider: providerRecord{
			APIKey:           "anthropic-secret",
			AnthropicBaseURL: server.URL,
		},
	}, capabilityTestPrompt, capabilityTestMaxOutputTokens)
	if err != nil {
		t.Fatal(err)
	}
	if out.Text != "ok" {
		t.Fatalf("text = %q", out.Text)
	}
}
