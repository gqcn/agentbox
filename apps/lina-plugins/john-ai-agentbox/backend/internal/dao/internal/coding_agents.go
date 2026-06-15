// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// CodingAgentsDao is the data access object for the table john_ai_agentbox_coding_agents.
type CodingAgentsDao struct {
	table    string              // table is the underlying table name of the DAO.
	group    string              // group is the database configuration group name of the current DAO.
	columns  CodingAgentsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler  // handlers for customized model modification.
}

// CodingAgentsColumns defines and stores column names for the table john_ai_agentbox_coding_agents.
type CodingAgentsColumns struct {
	Id            string // Agent ID
	UserId        string // Owner user ID
	Name          string // Agent display name
	ProviderId    string // Provider ID
	ModelName     string // Selected model name
	ModelProtocol string // Selected model protocol
	ImageId       string // Coding image ID
	AgentType     string // Agent runtime type
	IconKey       string // Agent icon key
	Notes         string // Agent notes
	DeletedAt     string // Soft deletion time
	CreatedAt     string // Creation time
	UpdatedAt     string // Update time
}

// codingAgentsColumns holds the columns for the table john_ai_agentbox_coding_agents.
var codingAgentsColumns = CodingAgentsColumns{
	Id:            "id",
	UserId:        "user_id",
	Name:          "name",
	ProviderId:    "provider_id",
	ModelName:     "model_name",
	ModelProtocol: "model_protocol",
	ImageId:       "image_id",
	AgentType:     "agent_type",
	IconKey:       "icon_key",
	Notes:         "notes",
	DeletedAt:     "deleted_at",
	CreatedAt:     "created_at",
	UpdatedAt:     "updated_at",
}

// NewCodingAgentsDao creates and returns a new DAO object for table data access.
func NewCodingAgentsDao(handlers ...gdb.ModelHandler) *CodingAgentsDao {
	return &CodingAgentsDao{
		group:    "default",
		table:    "john_ai_agentbox_coding_agents",
		columns:  codingAgentsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *CodingAgentsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *CodingAgentsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *CodingAgentsDao) Columns() CodingAgentsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *CodingAgentsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *CodingAgentsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *CodingAgentsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
