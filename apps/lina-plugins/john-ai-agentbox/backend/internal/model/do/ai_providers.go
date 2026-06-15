// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// AiProviders is the golang structure of table john_ai_agentbox_ai_providers for DAO operations like Where/Data.
type AiProviders struct {
	g.Meta           `orm:"table:john_ai_agentbox_ai_providers, do:true"`
	Id               any        // Provider ID
	Name             any        // Provider display name
	HomepageUrl      any        // Provider homepage URL
	Notes            any        // Provider notes
	ApiKey           any        // Provider API key
	OpenaiBaseUrl    any        // OpenAI-compatible base URL
	AnthropicBaseUrl any        // Anthropic-compatible base URL
	CreatedAt        *time.Time // Creation time
	UpdatedAt        *time.Time // Update time
}
