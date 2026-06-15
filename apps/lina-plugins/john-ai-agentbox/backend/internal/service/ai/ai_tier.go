// This file implements AI capability tier listing and updates. Tier responses
// are assembled with fixed-size batch queries for primary bindings and latest
// capability tests, avoiding per-tier database lookups.

package ai

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"

	"john-ai-agentbox/backend/internal/dao"
	"john-ai-agentbox/backend/internal/model/do"
	"john-ai-agentbox/backend/internal/model/entity"
)

type bindingProjectionRow struct {
	ID              int64      `orm:"id"`
	TierCode        string     `orm:"tier_code"`
	ProviderID      int64      `orm:"provider_id"`
	ProviderName    string     `orm:"provider_name"`
	ProviderModelID int64      `orm:"provider_model_id"`
	ModelName       string     `orm:"model_name"`
	Protocol        string     `orm:"protocol"`
	Priority        int        `orm:"priority"`
	Enabled         bool       `orm:"enabled"`
	CreatedAt       *time.Time `orm:"created_at"`
	UpdatedAt       *time.Time `orm:"updated_at"`
}

// ListTiers returns all fixed AgentBox AI capability tiers.
func (s *serviceImpl) ListTiers(ctx context.Context) ([]CapabilityTierInfo, error) {
	var rows []*entity.AiCapabilityTiers
	err := dao.AiCapabilityTiers.Ctx(ctx).Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeAIStoreUnavailable)
	}
	itemsByCode := make(map[string]CapabilityTierInfo, len(rows))
	for _, row := range rows {
		item := tierFromEntity(row)
		itemsByCode[item.Code] = item
	}
	bindings, err := primaryBindingsByTier(ctx)
	if err != nil {
		return nil, err
	}
	lastTests, err := latestCapabilityTestsByTier(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]CapabilityTierInfo, 0, len(rows))
	for _, code := range sortedTierCodes() {
		item, ok := itemsByCode[code]
		if !ok {
			continue
		}
		if binding := bindings[code]; binding != nil {
			item.Binding = binding
			item.Configured = binding.Enabled
			item.Available = item.Enabled && binding.Enabled
		}
		item.LastTest = lastTests[code]
		items = append(items, item)
	}
	return items, nil
}

// UpdateTier updates one AI capability tier and optionally its primary binding.
func (s *serviceImpl) UpdateTier(ctx context.Context, code string, input UpdateTierInput) (*CapabilityTierInfo, error) {
	code = normalizeTierCode(code)
	if code == "" {
		return nil, bizerr.NewCode(CodeAIInvalidInput)
	}
	if _, err := getTierEntity(ctx, code); err != nil {
		return nil, err
	}
	var (
		provider *providerRecord
		model    *providerModelRecord
	)
	if input.ProviderID > 0 || input.ProviderModelID > 0 {
		if input.ProviderID <= 0 || input.ProviderModelID <= 0 {
			return nil, bizerr.NewCode(CodeAIInvalidInput)
		}
		var err error
		provider, model, err = resolveProviderModel(ctx, input.ProviderID, input.ProviderModelID, input.Protocol)
		if err != nil {
			return nil, err
		}
	}
	err := dao.AiCapabilityTiers.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		_, updateErr := tx.Model(do.AiCapabilityTiers{}).Ctx(ctx).
			Where(do.AiCapabilityTiers{Code: code}).
			Data(do.AiCapabilityTiers{Enabled: input.Enabled}).
			Update()
		if updateErr != nil {
			return updateErr
		}
		if provider == nil || model == nil {
			return nil
		}
		var existing *entity.AiCapabilityBindings
		queryErr := tx.Model(do.AiCapabilityBindings{}).Ctx(ctx).
			Where(do.AiCapabilityBindings{
				TierCode: code,
				Priority: primaryBindingPriority,
			}).
			Scan(&existing)
		if queryErr != nil {
			return queryErr
		}
		data := do.AiCapabilityBindings{
			TierCode:        code,
			ProviderId:      provider.ID,
			ProviderModelId: model.ID,
			Priority:        primaryBindingPriority,
			Enabled:         true,
		}
		if existing != nil {
			_, queryErr = tx.Model(do.AiCapabilityBindings{}).Ctx(ctx).
				Where(do.AiCapabilityBindings{Id: existing.Id}).
				Data(data).
				Update()
			return queryErr
		}
		_, queryErr = tx.Model(do.AiCapabilityBindings{}).Ctx(ctx).
			Data(data).
			Insert()
		return queryErr
	})
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeAIStoreUnavailable)
	}
	return s.GetTier(ctx, code)
}

// GetTier returns one capability tier with current binding and latest test.
func (s *serviceImpl) GetTier(ctx context.Context, code string) (*CapabilityTierInfo, error) {
	row, err := getTierEntity(ctx, code)
	if err != nil {
		return nil, err
	}
	item := tierFromEntity(row)
	binding, err := primaryBinding(ctx, code)
	if err != nil {
		return nil, err
	}
	if binding != nil {
		item.Binding = binding
		item.Configured = binding.Enabled
		item.Available = item.Enabled && binding.Enabled
	}
	lastTests, err := latestCapabilityTestsByTier(ctx)
	if err != nil {
		return nil, err
	}
	item.LastTest = lastTests[code]
	return &item, nil
}

func getTierEntity(ctx context.Context, code string) (*entity.AiCapabilityTiers, error) {
	code = normalizeTierCode(code)
	if code == "" {
		return nil, bizerr.NewCode(CodeAIInvalidInput)
	}
	var row *entity.AiCapabilityTiers
	err := dao.AiCapabilityTiers.Ctx(ctx).
		Where(do.AiCapabilityTiers{Code: code}).
		Scan(&row)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeAIStoreUnavailable)
	}
	if row == nil {
		return nil, bizerr.NewCode(CodeAINotFound)
	}
	return row, nil
}

func tierFromEntity(row *entity.AiCapabilityTiers) CapabilityTierInfo {
	if row == nil {
		return CapabilityTierInfo{}
	}
	return CapabilityTierInfo{
		Code:        row.Code,
		DisplayName: row.DisplayName,
		Description: row.Description,
		Enabled:     row.Enabled,
		CreatedAt:   unixMilliFromTimePtr(row.CreatedAt),
		UpdatedAt:   unixMilliFromTimePtr(row.UpdatedAt),
	}
}

func primaryBinding(ctx context.Context, code string) (*CapabilityBindingInfo, error) {
	bindings, err := primaryBindingsByTier(ctx)
	if err != nil {
		return nil, err
	}
	return bindings[normalizeTierCode(code)], nil
}

func primaryBindingsByTier(ctx context.Context) (map[string]*CapabilityBindingInfo, error) {
	var rows []*bindingProjectionRow
	err := dao.AiCapabilityBindings.Ctx(ctx).
		As("b").
		Fields(`
			b.id,
			b.tier_code,
			b.provider_id,
			p.name AS provider_name,
			b.provider_model_id,
			m.name AS model_name,
			m.protocol,
			b.priority,
			b.enabled,
			b.created_at,
			b.updated_at
		`).
		LeftJoin(dao.AiProviders.Table()+" p", "p.id = b.provider_id").
		LeftJoin(dao.ProviderModels.Table()+" m", "m.id = b.provider_model_id").
		Where("b.priority", primaryBindingPriority).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeAIStoreUnavailable)
	}
	out := make(map[string]*CapabilityBindingInfo, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		out[row.TierCode] = &CapabilityBindingInfo{
			ID:              row.ID,
			TierCode:        row.TierCode,
			ProviderID:      row.ProviderID,
			ProviderName:    row.ProviderName,
			ProviderModelID: row.ProviderModelID,
			ModelName:       row.ModelName,
			Protocol:        row.Protocol,
			Priority:        row.Priority,
			Enabled:         row.Enabled,
			CreatedAt:       unixMilliFromTimePtr(row.CreatedAt),
			UpdatedAt:       unixMilliFromTimePtr(row.UpdatedAt),
		}
	}
	return out, nil
}

func latestCapabilityTestsByTier(ctx context.Context) (map[string]*InvocationLogInfo, error) {
	cols := dao.AiInvocationLogs.Columns()
	var idRows []struct {
		ID int64 `orm:"id"`
	}
	err := dao.AiInvocationLogs.Ctx(ctx).
		Fields("MAX(" + cols.Id + ") AS id").
		Where(do.AiInvocationLogs{Purpose: PurposeCapabilityTest}).
		Group(cols.TierCode).
		Scan(&idRows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeAIStoreUnavailable)
	}
	ids := make([]int64, 0, len(idRows))
	for _, row := range idRows {
		if row.ID > 0 {
			ids = append(ids, row.ID)
		}
	}
	out := make(map[string]*InvocationLogInfo, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	var rows []*entity.AiInvocationLogs
	err = dao.AiInvocationLogs.Ctx(ctx).
		WhereIn(cols.Id, ids).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeAIStoreUnavailable)
	}
	items, err := invocationLogsFromEntities(ctx, rows)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		copyItem := item
		out[item.TierCode] = &copyItem
	}
	return out, nil
}
