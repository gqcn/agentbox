// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AgentChatMessagesDao is the data access object for the table john_ai_agentbox_agent_chat_messages.
type AgentChatMessagesDao struct {
	table    string                   // table is the underlying table name of the DAO.
	group    string                   // group is the database configuration group name of the current DAO.
	columns  AgentChatMessagesColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler       // handlers for customized model modification.
}

// AgentChatMessagesColumns defines and stores column names for the table john_ai_agentbox_agent_chat_messages.
type AgentChatMessagesColumns struct {
	Id        string // Chat message ID
	SessionId string // Chat session ID
	Sequence  string // Message sequence in the session
	Role      string // Message role
	Content   string // Message content
	Status    string // Message status
	Metadata  string // Message metadata JSON
	CreatedAt string // Creation time
	UpdatedAt string // Update time
}

// agentChatMessagesColumns holds the columns for the table john_ai_agentbox_agent_chat_messages.
var agentChatMessagesColumns = AgentChatMessagesColumns{
	Id:        "id",
	SessionId: "session_id",
	Sequence:  "sequence",
	Role:      "role",
	Content:   "content",
	Status:    "status",
	Metadata:  "metadata",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
}

// NewAgentChatMessagesDao creates and returns a new DAO object for table data access.
func NewAgentChatMessagesDao(handlers ...gdb.ModelHandler) *AgentChatMessagesDao {
	return &AgentChatMessagesDao{
		group:    "default",
		table:    "john_ai_agentbox_agent_chat_messages",
		columns:  agentChatMessagesColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AgentChatMessagesDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AgentChatMessagesDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AgentChatMessagesDao) Columns() AgentChatMessagesColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AgentChatMessagesDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AgentChatMessagesDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AgentChatMessagesDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
