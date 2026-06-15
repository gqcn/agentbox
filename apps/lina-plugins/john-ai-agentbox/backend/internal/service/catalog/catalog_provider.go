// This file implements AgentBox provider persistence and model projection
// assembly. List responses batch-load provider models for all returned
// providers to avoid per-provider database queries.

package catalog

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"

	"john-ai-agentbox/backend/internal/dao"
	"john-ai-agentbox/backend/internal/model/do"
	"john-ai-agentbox/backend/internal/model/entity"
)

// ListProviders returns all providers with batched model projections.
func (s *serviceImpl) ListProviders(ctx context.Context) ([]ProviderInfo, error) {
	var rows []*entity.AiProviders
	err := dao.AiProviders.Ctx(ctx).
		OrderDesc(dao.AiProviders.Columns().Id).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeCatalogStoreUnavailable)
	}
	items := make([]ProviderInfo, 0, len(rows))
	providerIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		record := providerRecordFromEntity(row)
		items = append(items, record.ProviderInfo)
		providerIDs = append(providerIDs, record.ID)
	}
	modelsByProvider, err := s.listModelsByProviderIDs(ctx, providerIDs)
	if err != nil {
		return nil, err
	}
	for index := range items {
		items[index].Models = modelsByProvider[items[index].ID]
	}
	return items, nil
}

// CreateProvider creates one provider configuration in plugin-owned storage.
func (s *serviceImpl) CreateProvider(ctx context.Context, input ProviderInput) (*ProviderInfo, error) {
	normalized := normalizeProviderInput(input)
	if normalized.Name == "" {
		return nil, bizerr.NewCode(CodeCatalogInvalidInput)
	}
	id, err := dao.AiProviders.Ctx(ctx).Data(do.AiProviders{
		Name:             normalized.Name,
		HomepageUrl:      normalized.HomepageURL,
		Notes:            normalized.Notes,
		ApiKey:           normalized.APIKey,
		OpenaiBaseUrl:    normalized.OpenAIBaseURL,
		AnthropicBaseUrl: normalized.AnthropicBaseURL,
	}).InsertAndGetId()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeCatalogStoreUnavailable)
	}
	return s.GetProvider(ctx, id)
}

// GetProvider returns one provider and its model projections.
func (s *serviceImpl) GetProvider(ctx context.Context, id int64) (*ProviderInfo, error) {
	record, err := s.getProviderRecord(ctx, id)
	if err != nil {
		return nil, err
	}
	models, err := s.listProviderModels(ctx, id)
	if err != nil {
		return nil, err
	}
	info := record.ProviderInfo
	info.Models = models
	return &info, nil
}

// UpdateProvider updates one provider configuration.
func (s *serviceImpl) UpdateProvider(ctx context.Context, id int64, input ProviderInput) (*ProviderInfo, error) {
	normalized := normalizeProviderInput(input)
	if id <= 0 || normalized.Name == "" {
		return nil, bizerr.NewCode(CodeCatalogInvalidInput)
	}
	current, err := s.getProviderRecord(ctx, id)
	if err != nil {
		return nil, err
	}
	apiKey := normalized.APIKey
	if apiKey == "" {
		apiKey = current.APIKey
	}
	result, err := dao.AiProviders.Ctx(ctx).
		Where(do.AiProviders{Id: id}).
		Data(do.AiProviders{
			Name:             normalized.Name,
			HomepageUrl:      normalized.HomepageURL,
			Notes:            normalized.Notes,
			ApiKey:           apiKey,
			OpenaiBaseUrl:    normalized.OpenAIBaseURL,
			AnthropicBaseUrl: normalized.AnthropicBaseURL,
		}).
		Update()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeCatalogStoreUnavailable)
	}
	if affected, affectedErr := result.RowsAffected(); affectedErr == nil && affected == 0 {
		return nil, bizerr.NewCode(CodeCatalogNotFound)
	}
	return s.GetProvider(ctx, id)
}

// DeleteProvider deletes one unused provider configuration.
func (s *serviceImpl) DeleteProvider(ctx context.Context, id int64) error {
	if id <= 0 {
		return bizerr.NewCode(CodeCatalogInvalidInput)
	}
	if err := s.ensureProviderDeletable(ctx, id); err != nil {
		return err
	}
	result, err := dao.AiProviders.Ctx(ctx).
		Where(do.AiProviders{Id: id}).
		Delete()
	if err != nil {
		return bizerr.WrapCode(err, CodeCatalogStoreUnavailable)
	}
	if affected, affectedErr := result.RowsAffected(); affectedErr == nil && affected == 0 {
		return bizerr.NewCode(CodeCatalogNotFound)
	}
	return nil
}

// getProviderRecord reads one provider including secret material.
func (s *serviceImpl) getProviderRecord(ctx context.Context, id int64) (*ProviderRecord, error) {
	if id <= 0 {
		return nil, bizerr.NewCode(CodeCatalogInvalidInput)
	}
	var row *entity.AiProviders
	err := dao.AiProviders.Ctx(ctx).
		Where(do.AiProviders{Id: id}).
		Scan(&row)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeCatalogStoreUnavailable)
	}
	if row == nil {
		return nil, bizerr.NewCode(CodeCatalogNotFound)
	}
	record := providerRecordFromEntity(row)
	return &record, nil
}

// ensureProviderDeletable rejects deletion when provider-scoped data still references it.
func (s *serviceImpl) ensureProviderDeletable(ctx context.Context, id int64) error {
	if _, err := s.getProviderRecord(ctx, id); err != nil {
		return err
	}
	providerModelCount, err := dao.ProviderModels.Ctx(ctx).
		Where(do.ProviderModels{ProviderId: id}).
		Count()
	if err != nil {
		return bizerr.WrapCode(err, CodeCatalogStoreUnavailable)
	}
	if providerModelCount > 0 {
		return bizerr.NewCode(CodeCatalogResourceInUse)
	}
	agentCount, err := dao.CodingAgents.Ctx(ctx).
		Unscoped().
		Where(do.CodingAgents{ProviderId: id}).
		Count()
	if err != nil {
		return bizerr.WrapCode(err, CodeCatalogStoreUnavailable)
	}
	if agentCount > 0 {
		return bizerr.NewCode(CodeCatalogResourceInUse)
	}
	bindingCount, err := dao.AiCapabilityBindings.Ctx(ctx).
		Where(do.AiCapabilityBindings{ProviderId: id}).
		Count()
	if err != nil {
		return bizerr.WrapCode(err, CodeCatalogStoreUnavailable)
	}
	if bindingCount > 0 {
		return bizerr.NewCode(CodeCatalogResourceInUse)
	}
	return nil
}

// normalizeProviderInput trims provider strings before validation and storage.
func normalizeProviderInput(input ProviderInput) ProviderInput {
	return ProviderInput{
		Name:             strings.TrimSpace(input.Name),
		HomepageURL:      strings.TrimSpace(input.HomepageURL),
		Notes:            strings.TrimSpace(input.Notes),
		APIKey:           strings.TrimSpace(input.APIKey),
		OpenAIBaseURL:    strings.TrimSpace(input.OpenAIBaseURL),
		AnthropicBaseURL: strings.TrimSpace(input.AnthropicBaseURL),
	}
}

// providerRecordFromEntity maps generated provider entities into service projections.
func providerRecordFromEntity(row *entity.AiProviders) ProviderRecord {
	if row == nil {
		return ProviderRecord{}
	}
	record := ProviderRecord{
		ProviderInfo: ProviderInfo{
			ID:               row.Id,
			Name:             row.Name,
			HomepageURL:      row.HomepageUrl,
			Notes:            row.Notes,
			APIKeyMasked:     maskSecret(row.ApiKey),
			APIKeyConfigured: strings.TrimSpace(row.ApiKey) != "",
			OpenAIBaseURL:    row.OpenaiBaseUrl,
			AnthropicBaseURL: row.AnthropicBaseUrl,
			CreatedAt:        unixMilliFromTimePtr(row.CreatedAt),
			UpdatedAt:        unixMilliFromTimePtr(row.UpdatedAt),
		},
		APIKey: row.ApiKey,
	}
	return record
}

// maskSecret returns a stable non-reversible presentation form for secrets.
func maskSecret(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) <= 8 {
		return "****"
	}
	return value[:4] + "****" + value[len(value)-4:]
}

// unixMilliFromTimePtr converts nullable database timestamps for API projection.
func unixMilliFromTimePtr(value *time.Time) int64 {
	if value == nil || value.IsZero() {
		return 0
	}
	return value.UnixMilli()
}

// wrapCatalogStoreError converts low-level storage errors into structured bizerr values.
func wrapCatalogStoreError(err error) error {
	if err == nil {
		return nil
	}
	if bizerr.Is(err, CodeCatalogNotFound) ||
		bizerr.Is(err, CodeCatalogInvalidInput) ||
		bizerr.Is(err, CodeCatalogResourceInUse) ||
		bizerr.Is(err, CodeCatalogRemoteSyncFailed) {
		return err
	}
	if gerror.HasStack(err) {
		return bizerr.WrapCode(err, CodeCatalogStoreUnavailable)
	}
	return bizerr.WrapCode(err, CodeCatalogStoreUnavailable)
}
