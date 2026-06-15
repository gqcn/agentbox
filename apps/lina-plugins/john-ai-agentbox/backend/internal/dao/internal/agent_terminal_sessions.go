// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AgentTerminalSessionsDao is the data access object for the table john_ai_agentbox_agent_terminal_sessions.
type AgentTerminalSessionsDao struct {
	table    string                       // table is the underlying table name of the DAO.
	group    string                       // group is the database configuration group name of the current DAO.
	columns  AgentTerminalSessionsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler           // handlers for customized model modification.
}

// AgentTerminalSessionsColumns defines and stores column names for the table john_ai_agentbox_agent_terminal_sessions.
type AgentTerminalSessionsColumns struct {
	Id                 string // Terminal session ID
	UserId             string // Owner user ID
	AgentId            string // Agent ID
	TerminalId         string // Frontend terminal ID
	BackendType        string // Terminal backend type
	BackendSessionName string // Terminal backend session name
	WorkingDir         string // Working directory
	Shell              string // Shell path
	Status             string // Terminal session status
	LastError          string // Last terminal error
	ClosedAt           string // Close time
	CreatedAt          string // Creation time
	UpdatedAt          string // Update time
}

// agentTerminalSessionsColumns holds the columns for the table john_ai_agentbox_agent_terminal_sessions.
var agentTerminalSessionsColumns = AgentTerminalSessionsColumns{
	Id:                 "id",
	UserId:             "user_id",
	AgentId:            "agent_id",
	TerminalId:         "terminal_id",
	BackendType:        "backend_type",
	BackendSessionName: "backend_session_name",
	WorkingDir:         "working_dir",
	Shell:              "shell",
	Status:             "status",
	LastError:          "last_error",
	ClosedAt:           "closed_at",
	CreatedAt:          "created_at",
	UpdatedAt:          "updated_at",
}

// NewAgentTerminalSessionsDao creates and returns a new DAO object for table data access.
func NewAgentTerminalSessionsDao(handlers ...gdb.ModelHandler) *AgentTerminalSessionsDao {
	return &AgentTerminalSessionsDao{
		group:    "default",
		table:    "john_ai_agentbox_agent_terminal_sessions",
		columns:  agentTerminalSessionsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AgentTerminalSessionsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AgentTerminalSessionsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AgentTerminalSessionsDao) Columns() AgentTerminalSessionsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AgentTerminalSessionsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AgentTerminalSessionsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AgentTerminalSessionsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
