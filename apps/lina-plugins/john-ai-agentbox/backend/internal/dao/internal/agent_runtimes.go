// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AgentRuntimesDao is the data access object for the table john_ai_agentbox_agent_runtimes.
type AgentRuntimesDao struct {
	table    string               // table is the underlying table name of the DAO.
	group    string               // group is the database configuration group name of the current DAO.
	columns  AgentRuntimesColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler   // handlers for customized model modification.
}

// AgentRuntimesColumns defines and stores column names for the table john_ai_agentbox_agent_runtimes.
type AgentRuntimesColumns struct {
	AgentId         string // Agent ID
	ContainerId     string // AgentBox logical container ID
	DockerId        string // Docker container ID
	Status          string // Runtime status
	ConfigMountPath string // Runtime configuration mount path
	CreatedAt       string // Creation time
	UpdatedAt       string // Update time
}

// agentRuntimesColumns holds the columns for the table john_ai_agentbox_agent_runtimes.
var agentRuntimesColumns = AgentRuntimesColumns{
	AgentId:         "agent_id",
	ContainerId:     "container_id",
	DockerId:        "docker_id",
	Status:          "status",
	ConfigMountPath: "config_mount_path",
	CreatedAt:       "created_at",
	UpdatedAt:       "updated_at",
}

// NewAgentRuntimesDao creates and returns a new DAO object for table data access.
func NewAgentRuntimesDao(handlers ...gdb.ModelHandler) *AgentRuntimesDao {
	return &AgentRuntimesDao{
		group:    "default",
		table:    "john_ai_agentbox_agent_runtimes",
		columns:  agentRuntimesColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AgentRuntimesDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AgentRuntimesDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AgentRuntimesDao) Columns() AgentRuntimesColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AgentRuntimesDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AgentRuntimesDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AgentRuntimesDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
