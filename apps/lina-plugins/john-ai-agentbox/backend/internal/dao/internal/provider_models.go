// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ProviderModelsDao is the data access object for the table john_ai_agentbox_provider_models.
type ProviderModelsDao struct {
	table    string                // table is the underlying table name of the DAO.
	group    string                // group is the database configuration group name of the current DAO.
	columns  ProviderModelsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler    // handlers for customized model modification.
}

// ProviderModelsColumns defines and stores column names for the table john_ai_agentbox_provider_models.
type ProviderModelsColumns struct {
	Id           string // Provider model ID
	ProviderId   string // Provider ID
	Name         string // Model name
	Protocol     string // Model protocol
	Source       string // Model source
	LastSyncedAt string // Last remote synchronization time
	CreatedAt    string // Creation time
	UpdatedAt    string // Update time
}

// providerModelsColumns holds the columns for the table john_ai_agentbox_provider_models.
var providerModelsColumns = ProviderModelsColumns{
	Id:           "id",
	ProviderId:   "provider_id",
	Name:         "name",
	Protocol:     "protocol",
	Source:       "source",
	LastSyncedAt: "last_synced_at",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
}

// NewProviderModelsDao creates and returns a new DAO object for table data access.
func NewProviderModelsDao(handlers ...gdb.ModelHandler) *ProviderModelsDao {
	return &ProviderModelsDao{
		group:    "default",
		table:    "john_ai_agentbox_provider_models",
		columns:  providerModelsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ProviderModelsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ProviderModelsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ProviderModelsDao) Columns() ProviderModelsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ProviderModelsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ProviderModelsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ProviderModelsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
