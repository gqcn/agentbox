// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AgentChatInteractionsDao is the data access object for the table john_ai_agentbox_agent_chat_interactions.
type AgentChatInteractionsDao struct {
	table    string                       // table is the underlying table name of the DAO.
	group    string                       // group is the database configuration group name of the current DAO.
	columns  AgentChatInteractionsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler           // handlers for customized model modification.
}

// AgentChatInteractionsColumns defines and stores column names for the table john_ai_agentbox_agent_chat_interactions.
type AgentChatInteractionsColumns struct {
	Id                 string // Interaction ID
	AgentId            string // Agent ID
	SessionId          string // Chat session ID
	AssistantMessageId string // Related assistant message ID
	ToolType           string // Tool type
	ToolInteractionId  string // External tool interaction ID
	InteractionType    string // Interaction type
	Status             string // Interaction status
	Title              string // Interaction title
	Body               string // Interaction body
	RiskLevel          string // Risk level
	PayloadJson        string // Interaction payload JSON
	ResponseJson       string // Interaction response JSON
	ResponseMode       string // Response mode
	ResponseScope      string // Response scope
	ExpiresAt          string // Expiration time
	ResolvedAt         string // Resolution time
	CreatedAt          string // Creation time
	UpdatedAt          string // Update time
	DeletedAt          string // Soft deletion time
}

// agentChatInteractionsColumns holds the columns for the table john_ai_agentbox_agent_chat_interactions.
var agentChatInteractionsColumns = AgentChatInteractionsColumns{
	Id:                 "id",
	AgentId:            "agent_id",
	SessionId:          "session_id",
	AssistantMessageId: "assistant_message_id",
	ToolType:           "tool_type",
	ToolInteractionId:  "tool_interaction_id",
	InteractionType:    "interaction_type",
	Status:             "status",
	Title:              "title",
	Body:               "body",
	RiskLevel:          "risk_level",
	PayloadJson:        "payload_json",
	ResponseJson:       "response_json",
	ResponseMode:       "response_mode",
	ResponseScope:      "response_scope",
	ExpiresAt:          "expires_at",
	ResolvedAt:         "resolved_at",
	CreatedAt:          "created_at",
	UpdatedAt:          "updated_at",
	DeletedAt:          "deleted_at",
}

// NewAgentChatInteractionsDao creates and returns a new DAO object for table data access.
func NewAgentChatInteractionsDao(handlers ...gdb.ModelHandler) *AgentChatInteractionsDao {
	return &AgentChatInteractionsDao{
		group:    "default",
		table:    "john_ai_agentbox_agent_chat_interactions",
		columns:  agentChatInteractionsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AgentChatInteractionsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AgentChatInteractionsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AgentChatInteractionsDao) Columns() AgentChatInteractionsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AgentChatInteractionsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AgentChatInteractionsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AgentChatInteractionsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
