// This file defines version-one authentication DTOs for the AgentBox plugin.
// The route paths are plugin-relative; source-plugin registration publishes
// them under /x/john-ai-agentbox/api/v1.

package v1

import "github.com/gogf/gf/v2/frame/g"

// LoginReq creates a browser session from AgentBox credentials.
type LoginReq struct {
	g.Meta   `path:"/auth/sessions" method:"post" tags:"AgentBox Authentication" summary:"AgentBox login" dc:"Create an AgentBox browser session using username and password credentials, then set the independent HttpOnly agent_box_session cookie."`
	Username string `json:"username" v:"required" dc:"AgentBox login username" eg:"admin"`
	Password string `json:"password" v:"required" dc:"AgentBox login password" eg:"admin123"`
}

// LoginRes returns the authenticated AgentBox user.
type LoginRes = AuthSessionResponse

// SessionReq reads the current AgentBox browser session.
type SessionReq struct {
	g.Meta `path:"/auth/session" method:"get" tags:"AgentBox Authentication" summary:"Current AgentBox session" dc:"Read the AgentBox user associated with the independent HttpOnly agent_box_session cookie."`
}

// SessionRes returns the authenticated AgentBox user.
type SessionRes = AuthSessionResponse

// LogoutReq revokes the current AgentBox browser session.
type LogoutReq struct {
	g.Meta `path:"/auth/session" method:"delete" tags:"AgentBox Authentication" summary:"AgentBox logout" dc:"Revoke the current AgentBox server-side session and expire the independent agent_box_session cookie."`
}

// LogoutRes reports the logout result.
type LogoutRes = AuthLogoutResponse

// AuthSessionResponse is the public AgentBox session projection.
type AuthSessionResponse struct {
	User *AuthUser `json:"user" dc:"Authenticated AgentBox user, or omitted when there is no active session" eg:"{}"`
}

// AuthUser is the minimal current-user projection returned by auth endpoints.
type AuthUser struct {
	ID          string `json:"id" dc:"Stable AgentBox user ID" eg:"usr_admin"`
	Username    string `json:"username" dc:"AgentBox username" eg:"admin"`
	DisplayName string `json:"displayName" dc:"AgentBox user display name" eg:"Admin"`
	LastLoginAt int64  `json:"lastLoginAt" dc:"Last successful login time as Unix timestamp in milliseconds; 0 when unknown" eg:"1718000000000"`
}

// AuthLogoutResponse reports whether a session was revoked.
type AuthLogoutResponse struct {
	LoggedOut bool `json:"loggedOut" dc:"Whether the AgentBox session was revoked or already absent" eg:"true"`
}
