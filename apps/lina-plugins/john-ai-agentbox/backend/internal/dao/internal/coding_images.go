// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// CodingImagesDao is the data access object for the table john_ai_agentbox_coding_images.
type CodingImagesDao struct {
	table    string              // table is the underlying table name of the DAO.
	group    string              // group is the database configuration group name of the current DAO.
	columns  CodingImagesColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler  // handlers for customized model modification.
}

// CodingImagesColumns defines and stores column names for the table john_ai_agentbox_coding_images.
type CodingImagesColumns struct {
	Id           string // Coding image ID
	Name         string // Image display name
	ImageRef     string // Container image reference
	AgentType    string // Agent runtime type
	DefaultShell string // Default shell path
	Notes        string // Image notes
	Enabled      string // Whether the image is enabled
	IsDefault    string // Whether the image is a default option
	CreatedAt    string // Creation time
	UpdatedAt    string // Update time
}

// codingImagesColumns holds the columns for the table john_ai_agentbox_coding_images.
var codingImagesColumns = CodingImagesColumns{
	Id:           "id",
	Name:         "name",
	ImageRef:     "image_ref",
	AgentType:    "agent_type",
	DefaultShell: "default_shell",
	Notes:        "notes",
	Enabled:      "enabled",
	IsDefault:    "is_default",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
}

// NewCodingImagesDao creates and returns a new DAO object for table data access.
func NewCodingImagesDao(handlers ...gdb.ModelHandler) *CodingImagesDao {
	return &CodingImagesDao{
		group:    "default",
		table:    "john_ai_agentbox_coding_images",
		columns:  codingImagesColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *CodingImagesDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *CodingImagesDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *CodingImagesDao) Columns() CodingImagesColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *CodingImagesDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *CodingImagesDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *CodingImagesDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
