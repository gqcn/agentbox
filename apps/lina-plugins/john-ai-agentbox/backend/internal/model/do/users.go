// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Users is the golang structure of table john_ai_agentbox_users for DAO operations like Where/Data.
type Users struct {
	g.Meta       `orm:"table:john_ai_agentbox_users, do:true"`
	Id           any        // User ID
	Username     any        // Login username
	PasswordHash any        // BCrypt password hash
	Role         any        // AgentBox role
	Status       any        // User status
	LastLoginAt  *time.Time // Last successful login time
	DeletedAt    *time.Time // Soft deletion time
	CreatedAt    *time.Time // Creation time
	UpdatedAt    *time.Time // Update time
}
