// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// UserSessions is the golang structure of table john_ai_agentbox_user_sessions for DAO operations like Where/Data.
type UserSessions struct {
	g.Meta    `orm:"table:john_ai_agentbox_user_sessions, do:true"`
	TokenHash any        // Opaque session token hash
	UserId    any        // Owner user ID
	UserAgent any        // Client user agent
	IpAddress any        // Client IP address
	ExpiresAt *time.Time // Session expiration time
	RevokedAt *time.Time // Session revocation time
	CreatedAt *time.Time // Creation time
	UpdatedAt *time.Time // Update time
}
