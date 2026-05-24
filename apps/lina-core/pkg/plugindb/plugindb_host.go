//go:build !wasip1

// This file exposes the narrow host-side plugindb facade while keeping the
// concrete governance implementation under plugindb/internal.

package plugindb

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"

	plugindbhost "lina-core/pkg/plugindb/internal/host"
	plugindbplan "lina-core/pkg/plugindb/internal/plan"
)

// HostDB returns one governed host-side database wrapper for dynamic plugin data access.
func HostDB() (gdb.DB, error) {
	return plugindbhost.DB()
}

// AuditMetadata carries governed host-side execution metadata for audit and SQL checks.
type AuditMetadata = plugindbhost.AuditMetadata

// WithAudit attaches governed plugindb audit metadata to ctx.
func WithAudit(ctx context.Context, metadata *AuditMetadata) context.Context {
	return plugindbhost.WithAudit(ctx, metadata)
}

// DataQueryPlan represents one governed typed data query plan.
type DataQueryPlan = plugindbplan.DataQueryPlan

// DataFilter represents one governed typed filter clause.
type DataFilter = plugindbplan.DataFilter

// DataOrder represents one governed typed order clause.
type DataOrder = plugindbplan.DataOrder

// DataPagination represents one governed typed page window.
type DataPagination = plugindbplan.DataPagination

// DataPlanAction represents one governed data plan action.
type DataPlanAction = plugindbplan.DataPlanAction

// DataFilterOperator represents one governed filter operator.
type DataFilterOperator = plugindbplan.DataFilterOperator

// DataOrderDirection represents one governed order direction.
type DataOrderDirection = plugindbplan.DataOrderDirection

const (
	// DataPlanActionList lists records from one authorized table.
	DataPlanActionList = plugindbplan.DataPlanActionList
	// DataPlanActionGet reads one record by key from one authorized table.
	DataPlanActionGet = plugindbplan.DataPlanActionGet
	// DataPlanActionCount counts records from one authorized table.
	DataPlanActionCount = plugindbplan.DataPlanActionCount
	// DataOrderDirectionDESC orders records in descending order.
	DataOrderDirectionDESC = plugindbplan.DataOrderDirectionDESC
	// DataFilterOperatorEQ compares one field by equality.
	DataFilterOperatorEQ = plugindbplan.DataFilterOperatorEQ
	// DataFilterOperatorIN compares one field against a value list.
	DataFilterOperatorIN = plugindbplan.DataFilterOperatorIN
	// DataFilterOperatorLike compares one field by wildcard matching.
	DataFilterOperatorLike = plugindbplan.DataFilterOperatorLike
)

// UnmarshalQueryPlanJSON decodes one governed typed query plan.
func UnmarshalQueryPlanJSON(data []byte) (*DataQueryPlan, error) {
	return plugindbplan.UnmarshalQueryPlanJSON(data)
}

// ValidateDataQueryPlan validates one governed typed query plan.
func ValidateDataQueryPlan(plan *DataQueryPlan) error {
	return plugindbplan.ValidateDataQueryPlan(plan)
}

// ValidateDataFilter validates one governed typed filter clause.
func ValidateDataFilter(filter *DataFilter) error {
	return plugindbplan.ValidateDataFilter(filter)
}

// ValidateDataOrder validates one governed typed order clause.
func ValidateDataOrder(order *DataOrder) error {
	return plugindbplan.ValidateDataOrder(order)
}

// UnmarshalValueJSON decodes one JSON-encoded scalar or object value.
func UnmarshalValueJSON(data []byte) (any, error) {
	return plugindbplan.UnmarshalValueJSON(data)
}

// UnmarshalValuesJSON decodes one list of JSON-encoded values.
func UnmarshalValuesJSON(items [][]byte) ([]any, error) {
	return plugindbplan.UnmarshalValuesJSON(items)
}
