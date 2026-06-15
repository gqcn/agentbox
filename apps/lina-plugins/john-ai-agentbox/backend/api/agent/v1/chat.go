// This file defines Agent Chat DTOs for the AgentBox plugin. Paths are
// plugin-relative and are published under /x/john-ai-agentbox/api/v1 by source
// plugin route registration.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ChatSessionsReq lists chat sessions for one AgentBox agent.
type ChatSessionsReq struct {
	g.Meta `path:"/agents/{id}/chat/sessions" method:"get" tags:"AgentBox Chat" summary:"List AgentBox chat sessions" dc:"List Chat sessions owned by one authenticated-user-owned coding agent, ordered by last active time for the history panel."`
	ID     string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
}

// ChatSessionsRes returns chat sessions for one agent.
type ChatSessionsRes = []ChatSessionInfo

// CreateChatSessionReq creates one Chat session for one agent.
type CreateChatSessionReq struct {
	g.Meta `path:"/agents/{id}/chat/sessions" method:"post" tags:"AgentBox Chat" summary:"Create AgentBox chat session" dc:"Create a new empty Chat session for one authenticated-user-owned coding agent and return the session used by the Chat panel."`
	ID     string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
}

// CreateChatSessionRes returns the created Chat session.
type CreateChatSessionRes = ChatSessionInfo

// ChatSessionReq gets one Chat session for one agent.
type ChatSessionReq struct {
	g.Meta    `path:"/agents/{id}/chat/sessions/{sessionId}" method:"get" tags:"AgentBox Chat" summary:"Get AgentBox chat session" dc:"Get one Chat session after verifying it belongs to the authenticated-user-owned coding agent."`
	ID        string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	SessionID string `json:"sessionId" v:"required" dc:"Chat session ID that belongs to the agent" eg:"chat-1234567890abcdef1234567890abcdef"`
}

// ChatSessionRes returns one Chat session.
type ChatSessionRes = ChatSessionInfo

// UpdateChatSessionReq updates editable Chat session metadata.
type UpdateChatSessionReq struct {
	g.Meta    `path:"/agents/{id}/chat/sessions/{sessionId}" method:"put" tags:"AgentBox Chat" summary:"Update AgentBox chat session" dc:"Update editable metadata for one authenticated-user-owned Agent Chat session, currently the history-panel display title."`
	ID        string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	SessionID string `json:"sessionId" v:"required" dc:"Chat session ID that belongs to the agent" eg:"chat-1234567890abcdef1234567890abcdef"`
	Title     string `json:"title" v:"required" dc:"Chat session display title" eg:"修复文件上传失败"`
}

// UpdateChatSessionRes returns the updated Chat session.
type UpdateChatSessionRes = ChatSessionInfo

// DeleteChatSessionReq deletes one Chat session.
type DeleteChatSessionReq struct {
	g.Meta    `path:"/agents/{id}/chat/sessions/{sessionId}" method:"delete" tags:"AgentBox Chat" summary:"Delete AgentBox chat session" dc:"Delete one authenticated-user-owned Agent Chat session and cascade its persisted message and interaction history."`
	ID        string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	SessionID string `json:"sessionId" v:"required" dc:"Chat session ID that belongs to the agent" eg:"chat-1234567890abcdef1234567890abcdef"`
}

// DeleteChatSessionRes reports Chat session deletion state.
type DeleteChatSessionRes struct {
	Deleted bool `json:"deleted" dc:"Whether the Chat session was deleted" eg:"true"`
}

// ChatMessagesReq lists persisted Chat messages.
type ChatMessagesReq struct {
	g.Meta    `path:"/agents/{id}/chat/sessions/{sessionId}/messages" method:"get" tags:"AgentBox Chat" summary:"List AgentBox chat messages" dc:"List persisted Chat messages for one authenticated-user-owned Agent Chat session in sequence order."`
	ID        string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	SessionID string `json:"sessionId" v:"required" dc:"Chat session ID that belongs to the agent" eg:"chat-1234567890abcdef1234567890abcdef"`
}

// ChatMessagesRes returns session metadata and message history.
type ChatMessagesRes = ChatMessagesResponse

// ChatInteractionsReq lists interactions for one Chat session.
type ChatInteractionsReq struct {
	g.Meta    `path:"/agents/{id}/chat/sessions/{sessionId}/interactions" method:"get" tags:"AgentBox Chat" summary:"List AgentBox chat interactions" dc:"List persisted interaction requests for one authenticated-user-owned Chat session, optionally filtered by status or type."`
	ID        string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	SessionID string `json:"sessionId" v:"required" dc:"Chat session ID that belongs to the agent" eg:"chat-1234567890abcdef1234567890abcdef"`
	Status    string `json:"status" dc:"Filter by interaction status: pending, resolved, rejected, cancelled, expired, error; omitted means all statuses" eg:"pending"`
	Type      string `json:"type" dc:"Filter by interaction type: permission, question, choice, text, auth, plan, custom; omitted means all types" eg:"permission"`
}

// ChatInteractionsRes returns interactions for one Chat session.
type ChatInteractionsRes = []ChatInteractionInfo

// ChatInteractionReq gets one Chat interaction.
type ChatInteractionReq struct {
	g.Meta        `path:"/agents/{id}/chat/sessions/{sessionId}/interactions/{interactionId}" method:"get" tags:"AgentBox Chat" summary:"Get AgentBox chat interaction" dc:"Get one interaction request after verifying it belongs to the authenticated-user-owned Agent Chat session."`
	ID            string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	SessionID     string `json:"sessionId" v:"required" dc:"Chat session ID that belongs to the agent" eg:"chat-1234567890abcdef1234567890abcdef"`
	InteractionID string `json:"interactionId" v:"required" dc:"Chat interaction ID that belongs to the session" eg:"int-1234567890abcdef1234567890abcdef"`
}

// ChatInteractionRes returns one Chat interaction.
type ChatInteractionRes = ChatInteractionInfo

// UpdateChatInteractionResponseReq submits a user response.
type UpdateChatInteractionResponseReq struct {
	g.Meta        `path:"/agents/{id}/chat/sessions/{sessionId}/interactions/{interactionId}/response" method:"put" tags:"AgentBox Chat" summary:"Update AgentBox chat interaction response" dc:"Submit the user's structured response to a pending Chat interaction and persist the resolved or rejected interaction outcome."`
	ID            string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	SessionID     string `json:"sessionId" v:"required" dc:"Chat session ID that belongs to the agent" eg:"chat-1234567890abcdef1234567890abcdef"`
	InteractionID string `json:"interactionId" v:"required" dc:"Chat interaction ID that belongs to the session" eg:"int-1234567890abcdef1234567890abcdef"`
	Response      string `json:"response" v:"required" dc:"Structured user response as a JSON object string" eg:"{\"decision\":\"allow\"}"`
	ResponseMode  string `json:"responseMode" dc:"Response mode selected by the user: allow, answer, reject, cancel, allow_once, allow_session" eg:"allow"`
	ResponseScope string `json:"responseScope" dc:"Response scope: once, session, agent, provider; empty means no reusable scope" eg:"once"`
}

// UpdateChatInteractionResponseRes returns the updated interaction.
type UpdateChatInteractionResponseRes = ChatInteractionInfo

// UpdateChatInteractionStatusReq updates one pending interaction status.
type UpdateChatInteractionStatusReq struct {
	g.Meta        `path:"/agents/{id}/chat/sessions/{sessionId}/interactions/{interactionId}/status" method:"put" tags:"AgentBox Chat" summary:"Update AgentBox chat interaction status" dc:"Update a pending Chat interaction to a terminal status such as cancelled, expired, or error."`
	ID            string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	SessionID     string `json:"sessionId" v:"required" dc:"Chat session ID that belongs to the agent" eg:"chat-1234567890abcdef1234567890abcdef"`
	InteractionID string `json:"interactionId" v:"required" dc:"Chat interaction ID that belongs to the session" eg:"int-1234567890abcdef1234567890abcdef"`
	Status        string `json:"status" v:"required" dc:"Target interaction status: cancelled, expired, error" eg:"cancelled"`
}

// UpdateChatInteractionStatusRes returns the updated interaction.
type UpdateChatInteractionStatusRes = ChatInteractionInfo

// RecoverChatReq starts Chat recovery for one session.
type RecoverChatReq struct {
	g.Meta    `path:"/agents/{id}/chat/sessions/{sessionId}/recover" method:"post" tags:"AgentBox Chat" summary:"Recover AgentBox chat" dc:"Recover a specified Chat session after validating authenticated-user ownership; runtime startup is completed by the Chat runtime migration."`
	ID        string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	SessionID string `json:"sessionId" v:"required" dc:"Chat session ID that belongs to the agent" eg:"chat-1234567890abcdef1234567890abcdef"`
	StartNew  bool   `json:"startNew" dc:"Whether to start a new tool process from history when runtime recovery is available" eg:"true"`
}

// RecoverChatRes returns recovery state.
type RecoverChatRes = ChatRecoverResponse

// ChatSessionInfo is the public AgentBox Chat session projection.
type ChatSessionInfo struct {
	ID                 string `json:"id" dc:"Chat session ID" eg:"chat-1234567890abcdef1234567890abcdef"`
	AgentID            string `json:"agentId" dc:"Agent ID that owns this Chat session" eg:"agt-1234567890abcdef"`
	Title              string `json:"title" dc:"Chat session display title shown in the history panel" eg:"修复文件上传失败"`
	Status             string `json:"status" dc:"Session status: idle, running, waiting_input, exited, recovering, error" eg:"idle"`
	ToolType           string `json:"toolType" dc:"Abstract coding tool type bound to this Chat session" eg:"codex"`
	ToolSessionID      string `json:"toolSessionId,omitempty" dc:"Underlying tool session identifier, empty when no tool process is attached" eg:"codex-thread-123"`
	RuntimeState       string `json:"runtimeState" dc:"Runtime state: idle, running, waiting_input, exited, recovering, error" eg:"idle"`
	LastError          string `json:"lastError,omitempty" dc:"Last runtime error message, omitted when the session is healthy" eg:"process exited"`
	MessageCount       int64  `json:"messageCount" dc:"Persisted message count for this Chat session" eg:"12"`
	LastMessagePreview string `json:"lastMessagePreview" dc:"Preview text for the most recent message, empty when the session has no messages" eg:"已完成修复并通过测试"`
	CreatedAt          int64  `json:"createdAt" dc:"Session creation time as Unix timestamp in milliseconds" eg:"1704067200000"`
	UpdatedAt          int64  `json:"updatedAt" dc:"Session last update time as Unix timestamp in milliseconds" eg:"1704067300000"`
	LastActiveAt       int64  `json:"lastActiveAt" dc:"Session last activity time as Unix timestamp in milliseconds" eg:"1704067400000"`
}

// ChatMessageInfo is the public persisted Chat message projection.
type ChatMessageInfo struct {
	ID        int64  `json:"id" dc:"Chat message numeric ID" eg:"1001"`
	SessionID string `json:"sessionId" dc:"Chat session ID that owns this message" eg:"chat-1234567890abcdef1234567890abcdef"`
	Sequence  int64  `json:"sequence" dc:"Message sequence within the Chat session" eg:"3"`
	Role      string `json:"role" dc:"Message role: user, assistant, system, error, terminal" eg:"assistant"`
	Content   string `json:"content" dc:"Message content text" eg:"已完成修复并通过测试"`
	Status    string `json:"status" dc:"Message status: streaming, complete, error" eg:"complete"`
	Metadata  string `json:"metadata,omitempty" dc:"Optional JSON metadata string for tool-specific message details" eg:"{}"`
	CreatedAt int64  `json:"createdAt" dc:"Message creation time as Unix timestamp in milliseconds" eg:"1704067200000"`
	UpdatedAt int64  `json:"updatedAt" dc:"Message last update time as Unix timestamp in milliseconds" eg:"1704067300000"`
}

// ChatMessagesResponse returns a Chat session and ordered message history.
type ChatMessagesResponse struct {
	Session  ChatSessionInfo   `json:"session" dc:"Chat session metadata and history-list projection" eg:"{}"`
	Messages []ChatMessageInfo `json:"messages" dc:"Persisted messages ordered by sequence ascending" eg:"[]"`
}

// ChatInteractionInfo is the public Chat interaction projection.
type ChatInteractionInfo struct {
	ID                 string `json:"id" dc:"Chat interaction ID" eg:"int-1234567890abcdef1234567890abcdef"`
	AgentID            string `json:"agentId" dc:"Agent ID that owns this interaction" eg:"agt-1234567890abcdef"`
	SessionID          string `json:"sessionId" dc:"Chat session ID that owns this interaction" eg:"chat-1234567890abcdef1234567890abcdef"`
	AssistantMessageID int64  `json:"assistantMessageId,omitempty" dc:"Assistant message ID associated with this interaction" eg:"1001"`
	ToolType           string `json:"toolType" dc:"Abstract coding tool type that requested the interaction" eg:"claude_code"`
	ToolInteractionID  string `json:"toolInteractionId,omitempty" dc:"Underlying tool interaction identifier when available" eg:"toolu-123"`
	Type               string `json:"type" dc:"Interaction type: permission, question, choice, text, auth, plan, custom" eg:"permission"`
	Status             string `json:"status" dc:"Interaction status: pending, resolved, rejected, cancelled, expired, error" eg:"pending"`
	Title              string `json:"title" dc:"Short user-facing interaction title" eg:"允许执行 Bash 命令"`
	Body               string `json:"body" dc:"Detailed user-facing interaction prompt or explanation" eg:"Claude Code 请求执行测试命令"`
	RiskLevel          string `json:"riskLevel" dc:"Risk level: low, medium, high, critical" eg:"medium"`
	Payload            string `json:"payload" dc:"Structured interaction payload as JSON object string" eg:"{\"toolName\":\"Bash\"}"`
	Response           string `json:"response" dc:"Structured user response as JSON object string" eg:"{\"decision\":\"allow\"}"`
	ResponseMode       string `json:"responseMode,omitempty" dc:"Response mode selected by the user or adapter" eg:"allow"`
	ResponseScope      string `json:"responseScope,omitempty" dc:"Response scope: once, session, agent, provider" eg:"once"`
	ExpiresAt          *int64 `json:"expiresAt,omitempty" dc:"Interaction expiration time as Unix timestamp in milliseconds" eg:"1704067500000"`
	ResolvedAt         *int64 `json:"resolvedAt,omitempty" dc:"Interaction resolution time as Unix timestamp in milliseconds" eg:"1704067600000"`
	CreatedAt          int64  `json:"createdAt" dc:"Interaction creation time as Unix timestamp in milliseconds" eg:"1704067200000"`
	UpdatedAt          int64  `json:"updatedAt" dc:"Interaction last update time as Unix timestamp in milliseconds" eg:"1704067300000"`
}

// ChatRecoverResponse returns recovery state.
type ChatRecoverResponse struct {
	Session *ChatSessionInfo `json:"session,omitempty" dc:"Updated Chat session after recovery starts" eg:"{}"`
	Message *ChatMessageInfo `json:"message,omitempty" dc:"System message recorded for the recovery action" eg:"{}"`
}
