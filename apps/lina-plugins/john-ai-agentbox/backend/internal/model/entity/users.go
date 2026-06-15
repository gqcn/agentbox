// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// Users is the golang structure for table users.
type Users struct {
	Id           string     `json:"id"           orm:"id"            description:"User ID"`
	Username     string     `json:"username"     orm:"username"      description:"Login username"`
	PasswordHash string     `json:"passwordHash" orm:"password_hash" description:"BCrypt password hash"`
	Role         string     `json:"role"         orm:"role"          description:"AgentBox role"`
	Status       string     `json:"status"       orm:"status"        description:"User status"`
	LastLoginAt  *time.Time `json:"lastLoginAt"  orm:"last_login_at" description:"Last successful login time"`
	DeletedAt    *time.Time `json:"deletedAt"    orm:"deleted_at"    description:"Soft deletion time"`
	CreatedAt    *time.Time `json:"createdAt"    orm:"created_at"    description:"Creation time"`
	UpdatedAt    *time.Time `json:"updatedAt"    orm:"updated_at"    description:"Update time"`
}
