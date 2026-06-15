// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AgentChatSessionsDao is the data access object for the table john_ai_agentbox_agent_chat_sessions.
type AgentChatSessionsDao struct {
	table    string                   // table is the underlying table name of the DAO.
	group    string                   // group is the database configuration group name of the current DAO.
	columns  AgentChatSessionsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler       // handlers for customized model modification.
}

// AgentChatSessionsColumns defines and stores column names for the table john_ai_agentbox_agent_chat_sessions.
type AgentChatSessionsColumns struct {
	Id                 string // Chat session ID
	AgentId            string // Agent ID
	Title              string // Session title
	Status             string // Session status
	ToolType           string // Connected tool type
	ToolSessionId      string // Connected tool session ID
	RuntimeState       string // Runtime state
	LastError          string // Last runtime error
	MessageCount       string // Message count
	LastMessagePreview string // Latest message preview
	CreatedAt          string // Creation time
	UpdatedAt          string // Update time
	LastActiveAt       string // Last activity time
}

// agentChatSessionsColumns holds the columns for the table john_ai_agentbox_agent_chat_sessions.
var agentChatSessionsColumns = AgentChatSessionsColumns{
	Id:                 "id",
	AgentId:            "agent_id",
	Title:              "title",
	Status:             "status",
	ToolType:           "tool_type",
	ToolSessionId:      "tool_session_id",
	RuntimeState:       "runtime_state",
	LastError:          "last_error",
	MessageCount:       "message_count",
	LastMessagePreview: "last_message_preview",
	CreatedAt:          "created_at",
	UpdatedAt:          "updated_at",
	LastActiveAt:       "last_active_at",
}

// NewAgentChatSessionsDao creates and returns a new DAO object for table data access.
func NewAgentChatSessionsDao(handlers ...gdb.ModelHandler) *AgentChatSessionsDao {
	return &AgentChatSessionsDao{
		group:    "default",
		table:    "john_ai_agentbox_agent_chat_sessions",
		columns:  agentChatSessionsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AgentChatSessionsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AgentChatSessionsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AgentChatSessionsDao) Columns() AgentChatSessionsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AgentChatSessionsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AgentChatSessionsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AgentChatSessionsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
