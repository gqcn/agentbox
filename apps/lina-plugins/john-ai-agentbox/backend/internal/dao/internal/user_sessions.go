// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UserSessionsDao is the data access object for the table john_ai_agentbox_user_sessions.
type UserSessionsDao struct {
	table    string              // table is the underlying table name of the DAO.
	group    string              // group is the database configuration group name of the current DAO.
	columns  UserSessionsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler  // handlers for customized model modification.
}

// UserSessionsColumns defines and stores column names for the table john_ai_agentbox_user_sessions.
type UserSessionsColumns struct {
	TokenHash string // Opaque session token hash
	UserId    string // Owner user ID
	UserAgent string // Client user agent
	IpAddress string // Client IP address
	ExpiresAt string // Session expiration time
	RevokedAt string // Session revocation time
	CreatedAt string // Creation time
	UpdatedAt string // Update time
}

// userSessionsColumns holds the columns for the table john_ai_agentbox_user_sessions.
var userSessionsColumns = UserSessionsColumns{
	TokenHash: "token_hash",
	UserId:    "user_id",
	UserAgent: "user_agent",
	IpAddress: "ip_address",
	ExpiresAt: "expires_at",
	RevokedAt: "revoked_at",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
}

// NewUserSessionsDao creates and returns a new DAO object for table data access.
func NewUserSessionsDao(handlers ...gdb.ModelHandler) *UserSessionsDao {
	return &UserSessionsDao{
		group:    "default",
		table:    "john_ai_agentbox_user_sessions",
		columns:  userSessionsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *UserSessionsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *UserSessionsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *UserSessionsDao) Columns() UserSessionsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *UserSessionsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *UserSessionsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *UserSessionsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
