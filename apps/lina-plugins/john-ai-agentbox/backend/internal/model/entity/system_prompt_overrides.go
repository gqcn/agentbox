// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// SystemPromptOverrides is the golang structure for table system_prompt_overrides.
type SystemPromptOverrides struct {
	Code      string     `json:"code"      orm:"code"       description:"Prompt override code"`
	Content   string     `json:"content"   orm:"content"    description:"Prompt override content"`
	CreatedAt *time.Time `json:"createdAt" orm:"created_at" description:"Creation time"`
	UpdatedAt *time.Time `json:"updatedAt" orm:"updated_at" description:"Update time"`
}
