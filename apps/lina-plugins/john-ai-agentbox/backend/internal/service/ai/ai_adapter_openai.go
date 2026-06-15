// This file implements OpenAI-compatible non-streaming text generation for
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

type openAIChatRequest struct {
	Model       string              `json:"model"`
	Messages    []openAIChatMessage `json:"messages"`
	MaxTokens   int                 `json:"max_tokens,omitempty"`
	Temperature float64             `json:"temperature,omitempty"`
}

type openAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
			Refusal string `json:"refusal"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

func (s *serviceImpl) generateOpenAI(ctx context.Context, resolved ResolvedBinding, prompt string, maxOutputTokens int) (textGenerationOutput, error) {
	baseURL := providerBaseURL(resolved.Provider, catalog.ProtocolOpenAI)
	endpointURL, err := providerResourceURL(baseURL, openAIChatCompletionsPath)
	if err != nil {
		return textGenerationOutput{}, err
	}
	payload := openAIChatRequest{
		Model: resolved.Binding.ModelName,
		Messages: []openAIChatMessage{
			{Role: "system", Content: "Return only the requested text. Do not include explanations."},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   maxOutputTokens,
		Temperature: 0,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return textGenerationOutput{}, bizerrProviderFailed(gerror.Wrap(err, "encode OpenAI request"))
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpointURL, bytes.NewReader(body))
	if err != nil {
		return textGenerationOutput{}, bizerrProviderFailed(gerror.Wrap(err, "create OpenAI request"))
	}
	httpReq.Header.Set("Content-Type", "application/json")
	setProviderAPIKeyHeaders(httpReq, catalog.ProtocolOpenAI, resolved.Provider.APIKey)
	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return textGenerationOutput{}, bizerrProviderFailed(gerror.Wrap(err, "call OpenAI provider"))
	}
	if err := checkProviderResponse(resp, resolved.Provider.APIKey); err != nil {
		return textGenerationOutput{}, err
	}
	var payloadResp openAIChatResponse
	if err := decodeProviderJSON(resp, &payloadResp); err != nil {
		return textGenerationOutput{}, err
	}
	for _, choice := range payloadResp.Choices {
		if choice.Message.Content != "" {
			return textGenerationOutput{Text: choice.Message.Content}, nil
		}
	}
	return textGenerationOutput{}, bizerrProviderFailed(gerror.New("OpenAI response has empty message content"))
}
