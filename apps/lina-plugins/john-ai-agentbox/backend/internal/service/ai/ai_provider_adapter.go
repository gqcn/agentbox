// This file contains provider protocol adapter helpers shared by the
// OpenAI-compatible and Anthropic-compatible test calls.

package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"

	"john-ai-agentbox/backend/internal/service/catalog"
)

const (
	openAIChatCompletionsPath = "/v1/chat/completions"
	anthropicMessagesPath     = "/v1/messages"
	anthropicVersion          = "2023-06-01"
)

type textGenerationOutput struct {
	Text string
}

func (s *serviceImpl) generateWithBinding(ctx context.Context, resolved ResolvedBinding, prompt string, maxOutputTokens int) (textGenerationOutput, error) {
	switch resolved.Binding.Protocol {
	case catalog.ProtocolOpenAI:
		return s.generateOpenAI(ctx, resolved, prompt, maxOutputTokens)
	case catalog.ProtocolAnthropic:
		return s.generateAnthropic(ctx, resolved, prompt, maxOutputTokens)
	default:
		return textGenerationOutput{}, bizerrProviderFailed(gerror.New("ai model protocol is unsupported"))
	}
}

func checkProviderResponse(resp *http.Response, apiKey string) error {
	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return nil
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2048))
	closeErr := resp.Body.Close()
	if err != nil {
		return bizerrProviderFailed(err)
	}
	if closeErr != nil {
		return bizerrProviderFailed(closeErr)
	}
	message := strings.TrimSpace(string(body))
	if message == "" {
		message = resp.Status
	}
	return bizerrProviderFailed(gerror.Newf("ai provider request failed with status %d: %s", resp.StatusCode, sanitizeAIError(message, apiKey)))
}

func providerBaseURL(provider providerRecord, protocol string) string {
	if protocol == catalog.ProtocolAnthropic {
		return strings.TrimRight(strings.TrimSpace(provider.AnthropicBaseURL), "/")
	}
	return strings.TrimRight(strings.TrimSpace(provider.OpenAIBaseURL), "/")
}

func providerResourceURL(baseURL string, resourcePath string) (string, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return "", bizerrProviderFailed(gerror.New("provider base URL is required"))
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", bizerrProviderFailed(err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", bizerrProviderFailed(gerror.New("provider base URL must include scheme and host"))
	}
	currentPath := strings.TrimRight(parsed.EscapedPath(), "/")
	resourcePath = "/" + strings.Trim(strings.TrimSpace(resourcePath), "/")
	switch {
	case strings.HasSuffix(currentPath, resourcePath):
		return parsed.String(), nil
	case strings.HasSuffix(currentPath, "/chat/completions") && strings.HasSuffix(resourcePath, "/chat/completions"):
		return parsed.String(), nil
	case strings.HasSuffix(currentPath, "/messages") && strings.HasSuffix(resourcePath, "/messages"):
		return parsed.String(), nil
	case strings.HasSuffix(currentPath, "/v1") && strings.HasPrefix(resourcePath, "/v1/"):
		parsed.Path = path.Join(parsed.Path, strings.TrimPrefix(resourcePath, "/v1/"))
	default:
		parsed.Path = path.Join(parsed.Path, strings.TrimPrefix(resourcePath, "/"))
	}
	return parsed.String(), nil
}

func setProviderAPIKeyHeaders(req *http.Request, protocol string, apiKey string) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return
	}
	if protocol == catalog.ProtocolAnthropic {
		req.Header.Set("x-api-key", apiKey)
	} else {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	req.Header.Set("api-key", apiKey)
}

func decodeProviderJSON(resp *http.Response, target any) error {
	err := json.NewDecoder(resp.Body).Decode(target)
	closeErr := resp.Body.Close()
	if err != nil {
		return bizerrProviderFailed(err)
	}
	if closeErr != nil {
		return bizerrProviderFailed(closeErr)
	}
	return nil
}

func bizerrProviderFailed(err error) error {
	if err == nil {
		return nil
	}
	return bizerr.WrapCode(err, CodeAIProviderFailed)
}
