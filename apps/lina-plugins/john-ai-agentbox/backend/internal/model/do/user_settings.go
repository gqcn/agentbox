// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// UserSettings is the golang structure of table john_ai_agentbox_user_settings for DAO operations like Where/Data.
type UserSettings struct {
	g.Meta    `orm:"table:john_ai_agentbox_user_settings, do:true"`
	UserId    any        // Owner user ID
	Key       any        // Setting key
	Value     any        // Setting value
	CreatedAt *time.Time // Creation time
	UpdatedAt *time.Time // Update time
}
