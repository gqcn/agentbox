// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// UserSessions is the golang structure for table user_sessions.
type UserSessions struct {
	TokenHash string     `json:"tokenHash" orm:"token_hash" description:"Opaque session token hash"`
	UserId    string     `json:"userId"    orm:"user_id"    description:"Owner user ID"`
	UserAgent string     `json:"userAgent" orm:"user_agent" description:"Client user agent"`
	IpAddress string     `json:"ipAddress" orm:"ip_address" description:"Client IP address"`
	ExpiresAt *time.Time `json:"expiresAt" orm:"expires_at" description:"Session expiration time"`
	RevokedAt *time.Time `json:"revokedAt" orm:"revoked_at" description:"Session revocation time"`
	CreatedAt *time.Time `json:"createdAt" orm:"created_at" description:"Creation time"`
	UpdatedAt *time.Time `json:"updatedAt" orm:"updated_at" description:"Update time"`
}
