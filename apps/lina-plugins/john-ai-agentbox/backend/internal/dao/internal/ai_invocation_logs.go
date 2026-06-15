// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AiInvocationLogsDao is the data access object for the table john_ai_agentbox_ai_invocation_logs.
type AiInvocationLogsDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  AiInvocationLogsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// AiInvocationLogsColumns defines and stores column names for the table john_ai_agentbox_ai_invocation_logs.
type AiInvocationLogsColumns struct {
	Id              string // AI invocation log ID
	Purpose         string // Invocation purpose
	TierCode        string // Capability tier code
	ProviderId      string // Provider ID
	ProviderModelId string // Provider model ID
	ModelName       string // Model name used for the invocation
	Protocol        string // Model protocol
	Status          string // Invocation status
	LatencyMs       string // Invocation latency in milliseconds
	ErrorMessage    string // Invocation error message
	CreatedAt       string // Creation time
}

// aiInvocationLogsColumns holds the columns for the table john_ai_agentbox_ai_invocation_logs.
var aiInvocationLogsColumns = AiInvocationLogsColumns{
	Id:              "id",
	Purpose:         "purpose",
	TierCode:        "tier_code",
	ProviderId:      "provider_id",
	ProviderModelId: "provider_model_id",
	ModelName:       "model_name",
	Protocol:        "protocol",
	Status:          "status",
	LatencyMs:       "latency_ms",
	ErrorMessage:    "error_message",
	CreatedAt:       "created_at",
}

// NewAiInvocationLogsDao creates and returns a new DAO object for table data access.
func NewAiInvocationLogsDao(handlers ...gdb.ModelHandler) *AiInvocationLogsDao {
	return &AiInvocationLogsDao{
		group:    "default",
		table:    "john_ai_agentbox_ai_invocation_logs",
		columns:  aiInvocationLogsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AiInvocationLogsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AiInvocationLogsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AiInvocationLogsDao) Columns() AiInvocationLogsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AiInvocationLogsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AiInvocationLogsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AiInvocationLogsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
