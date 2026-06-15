// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AiProvidersDao is the data access object for the table john_ai_agentbox_ai_providers.
type AiProvidersDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  AiProvidersColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// AiProvidersColumns defines and stores column names for the table john_ai_agentbox_ai_providers.
type AiProvidersColumns struct {
	Id               string // Provider ID
	Name             string // Provider display name
	HomepageUrl      string // Provider homepage URL
	Notes            string // Provider notes
	ApiKey           string // Provider API key
	OpenaiBaseUrl    string // OpenAI-compatible base URL
	AnthropicBaseUrl string // Anthropic-compatible base URL
	CreatedAt        string // Creation time
	UpdatedAt        string // Update time
}

// aiProvidersColumns holds the columns for the table john_ai_agentbox_ai_providers.
var aiProvidersColumns = AiProvidersColumns{
	Id:               "id",
	Name:             "name",
	HomepageUrl:      "homepage_url",
	Notes:            "notes",
	ApiKey:           "api_key",
	OpenaiBaseUrl:    "openai_base_url",
	AnthropicBaseUrl: "anthropic_base_url",
	CreatedAt:        "created_at",
	UpdatedAt:        "updated_at",
}

// NewAiProvidersDao creates and returns a new DAO object for table data access.
func NewAiProvidersDao(handlers ...gdb.ModelHandler) *AiProvidersDao {
	return &AiProvidersDao{
		group:    "default",
		table:    "john_ai_agentbox_ai_providers",
		columns:  aiProvidersColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AiProvidersDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AiProvidersDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AiProvidersDao) Columns() AiProvidersColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AiProvidersDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AiProvidersDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AiProvidersDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
