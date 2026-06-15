// This file reads provider and provider-model records from plugin-owned tables.
// Secret material is kept in service-private structs and never returned through
// public HTTP response projections.

package ai

import (
	"context"

	"lina-core/pkg/bizerr"

	"john-ai-agentbox/backend/internal/dao"
	"john-ai-agentbox/backend/internal/model/do"
	"john-ai-agentbox/backend/internal/model/entity"
)

type providerRecord struct {
	ID               int64
	Name             string
	APIKey           string
	OpenAIBaseURL    string
	AnthropicBaseURL string
}

type providerModelRecord struct {
	ID         int64
	ProviderID int64
	Name       string
	Protocol   string
}

func getProviderRecord(ctx context.Context, id int64) (*providerRecord, error) {
	if id <= 0 {
		return nil, bizerr.NewCode(CodeAIInvalidInput)
	}
	var row *entity.AiProviders
	err := dao.AiProviders.Ctx(ctx).
		Where(do.AiProviders{Id: id}).
		Scan(&row)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeAIStoreUnavailable)
	}
	if row == nil {
		return nil, bizerr.NewCode(CodeAINotFound)
	}
	return &providerRecord{
		ID:               row.Id,
		Name:             row.Name,
		APIKey:           row.ApiKey,
		OpenAIBaseURL:    row.OpenaiBaseUrl,
		AnthropicBaseURL: row.AnthropicBaseUrl,
	}, nil
}

func getProviderModelRecord(ctx context.Context, id int64) (*providerModelRecord, error) {
	if id <= 0 {
		return nil, bizerr.NewCode(CodeAIInvalidInput)
	}
	var row *entity.ProviderModels
	err := dao.ProviderModels.Ctx(ctx).
		Where(do.ProviderModels{Id: id}).
		Scan(&row)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeAIStoreUnavailable)
	}
	if row == nil {
		return nil, bizerr.NewCode(CodeAINotFound)
	}
	return &providerModelRecord{
		ID:         row.Id,
		ProviderID: row.ProviderId,
		Name:       row.Name,
		Protocol:   row.Protocol,
	}, nil
}

func resolveProviderModel(ctx context.Context, providerID int64, modelID int64, protocol string) (*providerRecord, *providerModelRecord, error) {
	provider, err := getProviderRecord(ctx, providerID)
	if err != nil {
		return nil, nil, err
	}
	model, err := getProviderModelRecord(ctx, modelID)
	if err != nil {
		return nil, nil, err
	}
	if model.ProviderID != provider.ID {
		return nil, nil, bizerr.NewCode(CodeAINotFound)
	}
	targetProtocol := normalizeProtocol(protocol)
	if targetProtocol == "" {
		if protocol != "" {
			return nil, nil, bizerr.NewCode(CodeAIInvalidInput)
		}
		targetProtocol = model.Protocol
	}
	if model.Protocol != targetProtocol {
		return nil, nil, bizerr.NewCode(CodeAIInvalidInput)
	}
	if err := validateProviderProtocol(provider, targetProtocol); err != nil {
		return nil, nil, err
	}
	return provider, model, nil
}

func validateProviderProtocol(provider *providerRecord, protocol string) error {
	if provider == nil || normalizeProtocol(protocol) == "" {
		return bizerr.NewCode(CodeAIInvalidInput)
	}
	switch protocol {
	case "openai":
		if provider.OpenAIBaseURL == "" {
			return bizerr.NewCode(CodeAIInvalidInput)
		}
	case "anthropic":
		if provider.AnthropicBaseURL == "" {
			return bizerr.NewCode(CodeAIInvalidInput)
		}
	default:
		return bizerr.NewCode(CodeAIInvalidInput)
	}
	return nil
}

func providerNameMap(ctx context.Context, providerIDs []int64) (map[int64]string, error) {
	result := make(map[int64]string, len(providerIDs))
	if len(providerIDs) == 0 {
		return result, nil
	}
	cols := dao.AiProviders.Columns()
	var rows []*entity.AiProviders
	err := dao.AiProviders.Ctx(ctx).
		Fields(cols.Id, cols.Name).
		WhereIn(cols.Id, providerIDs).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeAIStoreUnavailable)
	}
	for _, row := range rows {
		result[row.Id] = row.Name
	}
	return result, nil
}
