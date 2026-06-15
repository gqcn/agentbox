// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AiCapabilityBindingsDao is the data access object for the table john_ai_agentbox_ai_capability_bindings.
type AiCapabilityBindingsDao struct {
	table    string                      // table is the underlying table name of the DAO.
	group    string                      // group is the database configuration group name of the current DAO.
	columns  AiCapabilityBindingsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler          // handlers for customized model modification.
}

// AiCapabilityBindingsColumns defines and stores column names for the table john_ai_agentbox_ai_capability_bindings.
type AiCapabilityBindingsColumns struct {
	Id              string // Capability binding ID
	TierCode        string // Capability tier code
	ProviderId      string // Provider ID
	ProviderModelId string // Provider model ID
	Priority        string // Binding priority
	Enabled         string // Whether the binding is enabled
	CreatedAt       string // Creation time
	UpdatedAt       string // Update time
}

// aiCapabilityBindingsColumns holds the columns for the table john_ai_agentbox_ai_capability_bindings.
var aiCapabilityBindingsColumns = AiCapabilityBindingsColumns{
	Id:              "id",
	TierCode:        "tier_code",
	ProviderId:      "provider_id",
	ProviderModelId: "provider_model_id",
	Priority:        "priority",
	Enabled:         "enabled",
	CreatedAt:       "created_at",
	UpdatedAt:       "updated_at",
}

// NewAiCapabilityBindingsDao creates and returns a new DAO object for table data access.
func NewAiCapabilityBindingsDao(handlers ...gdb.ModelHandler) *AiCapabilityBindingsDao {
	return &AiCapabilityBindingsDao{
		group:    "default",
		table:    "john_ai_agentbox_ai_capability_bindings",
		columns:  aiCapabilityBindingsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AiCapabilityBindingsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AiCapabilityBindingsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AiCapabilityBindingsDao) Columns() AiCapabilityBindingsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AiCapabilityBindingsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AiCapabilityBindingsDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rolls back the transaction and returns the error if function f returns a non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note: Do not commit or roll back the transaction in function f,
// as it is automatically handled by this function.
func (dao *AiCapabilityBindingsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
