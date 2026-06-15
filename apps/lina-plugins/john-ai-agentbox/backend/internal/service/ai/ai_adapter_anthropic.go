// This file implements Anthropic-compatible non-streaming text generation for
// AgentBox AI capability tests.

package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/gogf/gf/v2/errors/gerror"

	"john-ai-agentbox/backend/internal/service/catalog"
)

type anthropicMessagesRequest struct {
	Model       string                  `json:"model"`
	MaxTokens   int                     `json:"max_tokens"`
	Temperature float64                 `json:"temperature,omitempty"`
	System      string                  `json:"system,omitempty"`
	Messages    []anthropicMessageInput `json:"messages"`
}

type anthropicMessageInput struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicMessagesResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
}

func (s *serviceImpl) generateAnthropic(ctx context.Context, resolved ResolvedBinding, prompt string, maxOutputTokens int) (textGenerationOutput, error) {
	baseURL := providerBaseURL(resolved.Provider, catalog.ProtocolAnthropic)
	endpointURL, err := providerResourceURL(baseURL, anthropicMessagesPath)
	if err != nil {
		return textGenerationOutput{}, err
	}
	payload := anthropicMessagesRequest{
		Model:       resolved.Binding.ModelName,
		MaxTokens:   maxOutputTokens,
		Temperature: 0,
		System:      "Return only the requested text. Do not include explanations.",
		Messages: []anthropicMessageInput{
			{Role: "user", Content: prompt},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return textGenerationOutput{}, bizerrProviderFailed(gerror.Wrap(err, "encode Anthropic request"))
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpointURL, bytes.NewReader(body))
	if err != nil {
		return textGenerationOutput{}, bizerrProviderFailed(gerror.Wrap(err, "create Anthropic request"))
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("anthropic-version", anthropicVersion)
	setProviderAPIKeyHeaders(httpReq, catalog.ProtocolAnthropic, resolved.Provider.APIKey)
	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return textGenerationOutput{}, bizerrProviderFailed(gerror.Wrap(err, "call Anthropic provider"))
	}
	if err := checkProviderResponse(resp, resolved.Provider.APIKey); err != nil {
		return textGenerationOutput{}, err
	}
	var payloadResp anthropicMessagesResponse
	if err := decodeProviderJSON(resp, &payloadResp); err != nil {
		return textGenerationOutput{}, err
	}
	for _, part := range payloadResp.Content {
		if part.Type == "text" && part.Text != "" {
			return textGenerationOutput{Text: part.Text}, nil
		}
	}
	return textGenerationOutput{}, bizerrProviderFailed(gerror.New("Anthropic response has no text content"))
}
