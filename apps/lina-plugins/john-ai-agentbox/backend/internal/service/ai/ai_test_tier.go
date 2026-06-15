// This file implements lightweight AI capability tests. It records sanitized
// success or failure logs without persisting prompts or provider responses.

package ai

import (
	"context"
	"strings"
	"time"

	"lina-core/pkg/bizerr"
)

const (
	capabilityTestPrompt          = "Reply with exactly: ok"
	capabilityTestMaxOutputTokens = 32
)

// TestTier runs a lightweight prompt against a saved or draft tier binding.
func (s *serviceImpl) TestTier(ctx context.Context, code string, input TestTierInput) (*CapabilityTestResult, error) {
	code = normalizeTierCode(code)
	if code == "" {
		return nil, bizerr.NewCode(CodeAIInvalidInput)
	}
	started := time.Now()
	resolved, err := s.resolveTestBinding(ctx, code, input)
	if err != nil {
		log, logErr := createInvocationLog(ctx, invocationLogInput{
			Purpose:      PurposeCapabilityTest,
			TierCode:     code,
			Status:       InvocationStatusError,
			LatencyMS:    time.Since(started).Milliseconds(),
			ErrorMessage: err.Error(),
		})
		if logErr != nil {
			return nil, logErr
		}
		return testResultFromLog(*log, time.Now()), nil
	}
	testErr := s.runProviderTest(ctx, *resolved)
	status := InvocationStatusSuccess
	errorMessage := ""
	if testErr != nil {
		status = InvocationStatusError
		errorMessage = sanitizeAIError(testErr.Error(), resolved.Provider.APIKey)
	}
	log, err := createInvocationLog(ctx, invocationLogInput{
		Purpose:         PurposeCapabilityTest,
		TierCode:        code,
		ProviderID:      resolved.Binding.ProviderID,
		ProviderModelID: resolved.Binding.ProviderModelID,
		ModelName:       resolved.Binding.ModelName,
		Protocol:        resolved.Binding.Protocol,
		Status:          status,
		LatencyMS:       time.Since(started).Milliseconds(),
		ErrorMessage:    errorMessage,
	})
	if err != nil {
		return nil, err
	}
	return testResultFromLog(*log, time.Now()), nil
}

func (s *serviceImpl) resolveTestBinding(ctx context.Context, code string, input TestTierInput) (*ResolvedBinding, error) {
	if input.ProviderID > 0 || input.ProviderModelID > 0 {
		if input.ProviderID <= 0 || input.ProviderModelID <= 0 {
			return nil, bizerr.NewCode(CodeAIInvalidInput)
		}
		provider, model, err := resolveProviderModel(ctx, input.ProviderID, input.ProviderModelID, input.Protocol)
		if err != nil {
			return nil, err
		}
		return &ResolvedBinding{
			Tier: CapabilityTierInfo{
				Code:       code,
				Enabled:    true,
				Configured: true,
				Available:  true,
			},
			Binding: CapabilityBindingInfo{
				TierCode:        code,
				ProviderID:      provider.ID,
				ProviderName:    provider.Name,
				ProviderModelID: model.ID,
				ModelName:       model.Name,
				Protocol:        model.Protocol,
				Priority:        primaryBindingPriority,
				Enabled:         true,
			},
			Provider: *provider,
			Model:    *model,
		}, nil
	}
	tier, err := s.GetTier(ctx, code)
	if err != nil {
		return nil, err
	}
	if !tier.Enabled {
		return nil, bizerr.NewCode(CodeAIInvalidInput)
	}
	if tier.Binding == nil || !tier.Binding.Enabled {
		return nil, bizerr.NewCode(CodeAINotFound)
	}
	provider, model, err := resolveProviderModel(ctx, tier.Binding.ProviderID, tier.Binding.ProviderModelID, tier.Binding.Protocol)
	if err != nil {
		return nil, err
	}
	return &ResolvedBinding{
		Tier:     *tier,
		Binding:  *tier.Binding,
		Provider: *provider,
		Model:    *model,
	}, nil
}

func (s *serviceImpl) runProviderTest(ctx context.Context, resolved ResolvedBinding) error {
	output, err := s.generateWithBinding(ctx, resolved, capabilityTestPrompt, capabilityTestMaxOutputTokens)
	if err != nil {
		return err
	}
	if strings.TrimSpace(output.Text) == "" {
		return bizerr.NewCode(CodeAIProviderFailed)
	}
	return nil
}

func testResultFromLog(log InvocationLogInfo, testedAt time.Time) *CapabilityTestResult {
	return &CapabilityTestResult{
		Status:          log.Status,
		TierCode:        log.TierCode,
		ProviderID:      log.ProviderID,
		ProviderName:    log.ProviderName,
		ProviderModelID: log.ProviderModelID,
		ModelName:       log.ModelName,
		Protocol:        log.Protocol,
		LatencyMS:       log.LatencyMS,
		ErrorMessage:    log.ErrorMessage,
		TestedAt:        unixMilli(testedAt),
	}
}
