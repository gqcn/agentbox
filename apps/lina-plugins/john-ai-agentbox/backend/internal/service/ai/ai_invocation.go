// This file implements sanitized AgentBox AI invocation logs. Queries are
// database-filtered, bounded by a limit, and batch-load provider names to avoid
// per-row provider lookups.

package ai

import (
	"context"

	"lina-core/pkg/bizerr"

	"john-ai-agentbox/backend/internal/dao"
	"john-ai-agentbox/backend/internal/model/do"
	"john-ai-agentbox/backend/internal/model/entity"
)

type invocationLogInput struct {
	Purpose         string
	TierCode        string
	ProviderID      int64
	ProviderModelID int64
	ModelName       string
	Protocol        string
	Status          string
	LatencyMS       int64
	ErrorMessage    string
}

// ListInvocations returns sanitized AI invocation logs.
func (s *serviceImpl) ListInvocations(ctx context.Context, filter InvocationLogFilter) ([]InvocationLogInfo, error) {
	model := dao.AiInvocationLogs.Ctx(ctx)
	if filter.Purpose != "" {
		purpose := normalizePurpose(filter.Purpose)
		if purpose == "" {
			return nil, bizerr.NewCode(CodeAIInvalidInput)
		}
		model = model.Where(do.AiInvocationLogs{Purpose: purpose})
	}
	if filter.TierCode != "" {
		tierCode := normalizeTierCode(filter.TierCode)
		if tierCode == "" {
			return nil, bizerr.NewCode(CodeAIInvalidInput)
		}
		model = model.Where(do.AiInvocationLogs{TierCode: tierCode})
	}
	if filter.Status != "" {
		status := normalizeStatus(filter.Status)
		if status == "" {
			return nil, bizerr.NewCode(CodeAIInvalidInput)
		}
		model = model.Where(do.AiInvocationLogs{Status: status})
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = defaultLogLimit
	}
	if limit > maxLogLimit {
		limit = maxLogLimit
	}
	cols := dao.AiInvocationLogs.Columns()
	var rows []*entity.AiInvocationLogs
	err := model.
		OrderDesc(cols.CreatedAt).
		OrderDesc(cols.Id).
		Limit(limit).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeAIStoreUnavailable)
	}
	items, err := invocationLogsFromEntities(ctx, rows)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func createInvocationLog(ctx context.Context, input invocationLogInput) (*InvocationLogInfo, error) {
	purpose := normalizePurpose(input.Purpose)
	tierCode := normalizeTierCode(input.TierCode)
	status := normalizeStatus(input.Status)
	protocol := normalizeProtocol(input.Protocol)
	if purpose == "" || tierCode == "" || status == "" {
		return nil, bizerr.NewCode(CodeAIInvalidInput)
	}
	data := do.AiInvocationLogs{
		Purpose:      purpose,
		TierCode:     tierCode,
		ModelName:    input.ModelName,
		Protocol:     protocol,
		Status:       status,
		LatencyMs:    input.LatencyMS,
		ErrorMessage: sanitizeAIError(input.ErrorMessage, ""),
	}
	if input.ProviderID > 0 {
		data.ProviderId = input.ProviderID
	}
	if input.ProviderModelID > 0 {
		data.ProviderModelId = input.ProviderModelID
	}
	id, err := dao.AiInvocationLogs.Ctx(ctx).Data(data).InsertAndGetId()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeAIStoreUnavailable)
	}
	var row *entity.AiInvocationLogs
	err = dao.AiInvocationLogs.Ctx(ctx).
		Where(do.AiInvocationLogs{Id: id}).
		Scan(&row)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeAIStoreUnavailable)
	}
	items, err := invocationLogsFromEntities(ctx, []*entity.AiInvocationLogs{row})
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, bizerr.NewCode(CodeAINotFound)
	}
	return &items[0], nil
}

func invocationLogsFromEntities(ctx context.Context, rows []*entity.AiInvocationLogs) ([]InvocationLogInfo, error) {
	providerIDs := uniqueInvocationProviderIDs(rows)
	names, err := providerNameMap(ctx, providerIDs)
	if err != nil {
		return nil, err
	}
	items := make([]InvocationLogInfo, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		item := InvocationLogInfo{
			ID:              row.Id,
			Purpose:         row.Purpose,
			TierCode:        row.TierCode,
			ProviderID:      row.ProviderId,
			ProviderName:    names[row.ProviderId],
			ProviderModelID: row.ProviderModelId,
			ModelName:       row.ModelName,
			Protocol:        row.Protocol,
			Status:          row.Status,
			LatencyMS:       row.LatencyMs,
			ErrorMessage:    row.ErrorMessage,
			CreatedAt:       unixMilliFromTimePtr(row.CreatedAt),
		}
		items = append(items, item)
	}
	return items, nil
}

func uniqueInvocationProviderIDs(rows []*entity.AiInvocationLogs) []int64 {
	seen := make(map[int64]struct{}, len(rows))
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row == nil || row.ProviderId <= 0 {
			continue
		}
		if _, ok := seen[row.ProviderId]; ok {
			continue
		}
		seen[row.ProviderId] = struct{}{}
		ids = append(ids, row.ProviderId)
	}
	return ids
}
