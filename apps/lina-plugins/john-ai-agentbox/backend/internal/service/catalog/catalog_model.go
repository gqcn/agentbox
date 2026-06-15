// This file implements provider-model persistence and remote synchronization.
// Model list assembly is always batched by provider IDs; remote synchronization
// has an explicit per-call cap before writing plugin-owned records.

package catalog

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"

	"john-ai-agentbox/backend/internal/dao"
	"john-ai-agentbox/backend/internal/model/do"
	"john-ai-agentbox/backend/internal/model/entity"
)

const (
	openAIModelsPath      = "/v1/models"
	anthropicModelsPath   = "/v1/models"
	anthropicVersionValue = "2023-06-01"
)

// CreateProviderModel creates or updates one manually managed model record.
func (s *serviceImpl) CreateProviderModel(ctx context.Context, providerID int64, input ProviderModelInput) (*ProviderModelInfo, error) {
	item, err := s.upsertProviderModel(ctx, providerID, input, ModelSourceManual)
	if err != nil {
		return nil, err
	}
	return item, nil
}

// DeleteProviderModel deletes one unused provider model record.
func (s *serviceImpl) DeleteProviderModel(ctx context.Context, providerID int64, modelID int64) error {
	if providerID <= 0 || modelID <= 0 {
		return bizerr.NewCode(CodeCatalogInvalidInput)
	}
	model, err := s.getProviderModel(ctx, modelID)
	if err != nil {
		return err
	}
	if model.ProviderID != providerID {
		return bizerr.NewCode(CodeCatalogNotFound)
	}
	inUse, err := s.providerModelInUse(ctx, *model)
	if err != nil {
		return err
	}
	if inUse {
		return bizerr.NewCode(CodeCatalogResourceInUse)
	}
	result, err := dao.ProviderModels.Ctx(ctx).
		Where(do.ProviderModels{Id: modelID}).
		Delete()
	if err != nil {
		return bizerr.WrapCode(err, CodeCatalogStoreUnavailable)
	}
	if affected, affectedErr := result.RowsAffected(); affectedErr == nil && affected == 0 {
		return bizerr.NewCode(CodeCatalogNotFound)
	}
	return nil
}

// SyncProviderModels fetches and persists remote provider model IDs.
func (s *serviceImpl) SyncProviderModels(ctx context.Context, providerID int64, protocol string) (*SyncProviderModelsOutput, error) {
	provider, err := s.getProviderRecord(ctx, providerID)
	if err != nil {
		return nil, err
	}
	protocol = normalizeProtocol(protocol)
	if protocol == "" {
		return nil, bizerr.NewCode(CodeCatalogInvalidInput)
	}
	baseURL := provider.OpenAIBaseURL
	if protocol == ProtocolAnthropic {
		baseURL = provider.AnthropicBaseURL
	}
	modelsURL, err := providerModelsURL(baseURL, protocol)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeCatalogRemoteSyncFailed)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, modelsURL, nil)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeCatalogRemoteSyncFailed)
	}
	setProviderAPIKeyHeaders(req, protocol, provider.APIKey)
	if protocol == ProtocolAnthropic {
		req.Header.Set("anthropic-version", anthropicVersionValue)
	}
	response, err := s.httpClient.Do(req)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeCatalogRemoteSyncFailed)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, bizerr.NewCode(CodeCatalogRemoteSyncFailed)
	}
	var payload modelListResponse
	if err = json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, bizerr.WrapCode(err, CodeCatalogRemoteSyncFailed)
	}
	items := make([]ProviderModelInfo, 0, minInt(len(payload.Data), s.remoteModelSyncLimit))
	seen := make(map[string]struct{}, len(payload.Data))
	for _, row := range payload.Data {
		name := strings.TrimSpace(row.ID)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		if len(items) >= s.remoteModelSyncLimit {
			return nil, bizerr.NewCode(CodeCatalogRemoteSyncFailed)
		}
		saved, saveErr := s.upsertProviderModel(ctx, providerID, ProviderModelInput{
			Name:     name,
			Protocol: protocol,
		}, ModelSourceAPI)
		if saveErr != nil {
			return nil, saveErr
		}
		items = append(items, *saved)
	}
	return &SyncProviderModelsOutput{
		Protocol: protocol,
		Count:    len(items),
		Models:   items,
	}, nil
}

// upsertProviderModel creates or updates a provider model by its stable unique key.
func (s *serviceImpl) upsertProviderModel(ctx context.Context, providerID int64, input ProviderModelInput, source string) (*ProviderModelInfo, error) {
	if _, err := s.getProviderRecord(ctx, providerID); err != nil {
		return nil, err
	}
	name := strings.TrimSpace(input.Name)
	protocol := normalizeProtocol(input.Protocol)
	source = normalizeModelSource(source)
	if providerID <= 0 || name == "" || protocol == "" || source == "" {
		return nil, bizerr.NewCode(CodeCatalogInvalidInput)
	}

	var modelID int64
	err := dao.ProviderModels.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var existing *entity.ProviderModels
		queryErr := tx.Model(do.ProviderModels{}).Ctx(ctx).
			Where(do.ProviderModels{
				ProviderId: providerID,
				Name:       name,
				Protocol:   protocol,
			}).
			Scan(&existing)
		if queryErr != nil {
			return queryErr
		}
		if existing != nil {
			modelID = existing.Id
			updateData := do.ProviderModels{Source: source}
			if source == ModelSourceAPI {
				now := time.Now()
				updateData.LastSyncedAt = &now
			}
			_, queryErr = tx.Model(do.ProviderModels{}).Ctx(ctx).
				Where(do.ProviderModels{Id: modelID}).
				Data(updateData).
				Update()
			return queryErr
		}
		insertData := do.ProviderModels{
			ProviderId: providerID,
			Name:       name,
			Protocol:   protocol,
			Source:     source,
		}
		if source == ModelSourceAPI {
			now := time.Now()
			insertData.LastSyncedAt = &now
		}
		modelID, queryErr = tx.Model(do.ProviderModels{}).Ctx(ctx).
			Data(insertData).
			InsertAndGetId()
		return queryErr
	})
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeCatalogStoreUnavailable)
	}
	return s.getProviderModel(ctx, modelID)
}

// getProviderModel returns one provider model projection by ID.
func (s *serviceImpl) getProviderModel(ctx context.Context, id int64) (*ProviderModelInfo, error) {
	if id <= 0 {
		return nil, bizerr.NewCode(CodeCatalogInvalidInput)
	}
	var row *entity.ProviderModels
	err := dao.ProviderModels.Ctx(ctx).
		Where(do.ProviderModels{Id: id}).
		Scan(&row)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeCatalogStoreUnavailable)
	}
	if row == nil {
		return nil, bizerr.NewCode(CodeCatalogNotFound)
	}
	item := providerModelFromEntity(row)
	return &item, nil
}

// listProviderModels returns all model projections for one provider.
func (s *serviceImpl) listProviderModels(ctx context.Context, providerID int64) ([]ProviderModelInfo, error) {
	modelsByProvider, err := s.listModelsByProviderIDs(ctx, []int64{providerID})
	if err != nil {
		return nil, err
	}
	return modelsByProvider[providerID], nil
}

// listModelsByProviderIDs batch-loads provider models for current provider rows.
func (s *serviceImpl) listModelsByProviderIDs(ctx context.Context, providerIDs []int64) (map[int64][]ProviderModelInfo, error) {
	result := make(map[int64][]ProviderModelInfo, len(providerIDs))
	if len(providerIDs) == 0 {
		return result, nil
	}
	cols := dao.ProviderModels.Columns()
	var rows []*entity.ProviderModels
	err := dao.ProviderModels.Ctx(ctx).
		WhereIn(cols.ProviderId, providerIDs).
		OrderAsc(cols.ProviderId).
		OrderAsc(cols.Protocol).
		OrderAsc(cols.Name).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeCatalogStoreUnavailable)
	}
	for _, row := range rows {
		item := providerModelFromEntity(row)
		result[item.ProviderID] = append(result[item.ProviderID], item)
	}
	return result, nil
}

// providerModelInUse checks agent and AI capability references for one model.
func (s *serviceImpl) providerModelInUse(ctx context.Context, item ProviderModelInfo) (bool, error) {
	agentCount, err := dao.CodingAgents.Ctx(ctx).
		Unscoped().
		Where(do.CodingAgents{
			ProviderId:    item.ProviderID,
			ModelName:     item.Name,
			ModelProtocol: item.Protocol,
		}).
		Count()
	if err != nil {
		return false, bizerr.WrapCode(err, CodeCatalogStoreUnavailable)
	}
	if agentCount > 0 {
		return true, nil
	}
	bindingCount, err := dao.AiCapabilityBindings.Ctx(ctx).
		Where(do.AiCapabilityBindings{ProviderModelId: item.ID}).
		Count()
	if err != nil {
		return false, bizerr.WrapCode(err, CodeCatalogStoreUnavailable)
	}
	return bindingCount > 0, nil
}

// providerModelFromEntity maps generated model entities into service projections.
func providerModelFromEntity(row *entity.ProviderModels) ProviderModelInfo {
	if row == nil {
		return ProviderModelInfo{}
	}
	item := ProviderModelInfo{
		ID:         row.Id,
		ProviderID: row.ProviderId,
		Name:       row.Name,
		Protocol:   row.Protocol,
		Source:     row.Source,
		CreatedAt:  unixMilliFromTimePtr(row.CreatedAt),
		UpdatedAt:  unixMilliFromTimePtr(row.UpdatedAt),
	}
	if value := unixMilliFromTimePtr(row.LastSyncedAt); value > 0 {
		item.LastSyncedAt = &value
	}
	return item
}

// modelListResponse is the provider-neutral remote model-list JSON shape.
type modelListResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

// providerModelsURL resolves the concrete remote model-list endpoint.
func providerModelsURL(baseURL string, protocol string) (string, error) {
	if strings.TrimSpace(baseURL) == "" {
		return "", bizerr.NewCode(CodeCatalogInvalidInput)
	}
	if protocol == ProtocolAnthropic {
		return providerResourceURL(baseURL, anthropicModelsPath)
	}
	return providerResourceURL(baseURL, openAIModelsPath)
}

// providerResourceURL appends one provider resource path without duplicating /v1/models.
func providerResourceURL(baseURL string, resourcePath string) (string, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", bizerr.NewCode(CodeCatalogInvalidInput)
	}
	currentPath := strings.TrimRight(parsed.EscapedPath(), "/")
	resourcePath = "/" + strings.Trim(strings.TrimSpace(resourcePath), "/")
	switch {
	case strings.HasSuffix(currentPath, resourcePath):
		return parsed.String(), nil
	case strings.HasSuffix(currentPath, "/models") && strings.HasSuffix(resourcePath, "/models"):
		return parsed.String(), nil
	case strings.HasSuffix(currentPath, "/v1") && strings.HasPrefix(resourcePath, "/v1/"):
		parsed.Path = path.Join(parsed.Path, strings.TrimPrefix(resourcePath, "/v1/"))
	default:
		parsed.Path = path.Join(parsed.Path, strings.TrimPrefix(resourcePath, "/"))
	}
	return parsed.String(), nil
}

// setProviderAPIKeyHeaders sends protocol-specific and gateway-neutral API key headers.
func setProviderAPIKeyHeaders(req *http.Request, protocol string, apiKey string) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return
	}
	if protocol == ProtocolAnthropic {
		req.Header.Set("x-api-key", apiKey)
	} else {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	req.Header.Set("api-key", apiKey)
}

// normalizeProtocol constrains model protocol to supported values.
func normalizeProtocol(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case ProtocolOpenAI:
		return ProtocolOpenAI
	case ProtocolAnthropic:
		return ProtocolAnthropic
	default:
		return ""
	}
}

// normalizeModelSource constrains provider model source to supported values.
func normalizeModelSource(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case ModelSourceManual:
		return ModelSourceManual
	case ModelSourceAPI:
		return ModelSourceAPI
	default:
		return ""
	}
}

// minInt returns the smaller integer.
func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
