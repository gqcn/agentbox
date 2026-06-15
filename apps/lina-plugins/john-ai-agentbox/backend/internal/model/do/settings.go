// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Settings is the golang structure of table john_ai_agentbox_settings for DAO operations like Where/Data.
type Settings struct {
	g.Meta    `orm:"table:john_ai_agentbox_settings, do:true"`
	Key       any        // Setting key
	Value     any        // Setting value
	CreatedAt *time.Time // Creation time
	UpdatedAt *time.Time // Update time
}
