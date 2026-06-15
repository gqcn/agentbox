// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// UserSettings is the golang structure for table user_settings.
type UserSettings struct {
	UserId    string     `json:"userId"    orm:"user_id"    description:"Owner user ID"`
	Key       string     `json:"key"       orm:"key"        description:"Setting key"`
	Value     string     `json:"value"     orm:"value"      description:"Setting value"`
	CreatedAt *time.Time `json:"createdAt" orm:"created_at" description:"Creation time"`
	UpdatedAt *time.Time `json:"updatedAt" orm:"updated_at" description:"Update time"`
}
