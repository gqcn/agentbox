// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// AiProviders is the golang structure for table ai_providers.
type AiProviders struct {
	Id               int64      `json:"id"               orm:"id"                 description:"Provider ID"`
	Name             string     `json:"name"             orm:"name"               description:"Provider display name"`
	HomepageUrl      string     `json:"homepageUrl"      orm:"homepage_url"       description:"Provider homepage URL"`
	Notes            string     `json:"notes"            orm:"notes"              description:"Provider notes"`
	ApiKey           string     `json:"apiKey"           orm:"api_key"            description:"Provider API key"`
	OpenaiBaseUrl    string     `json:"openaiBaseUrl"    orm:"openai_base_url"    description:"OpenAI-compatible base URL"`
	AnthropicBaseUrl string     `json:"anthropicBaseUrl" orm:"anthropic_base_url" description:"Anthropic-compatible base URL"`
	CreatedAt        *time.Time `json:"createdAt"        orm:"created_at"         description:"Creation time"`
	UpdatedAt        *time.Time `json:"updatedAt"        orm:"updated_at"         description:"Update time"`
}
