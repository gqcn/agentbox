// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// SystemPromptOverrides is the golang structure of table john_ai_agentbox_system_prompt_overrides for DAO operations like Where/Data.
type SystemPromptOverrides struct {
	g.Meta    `orm:"table:john_ai_agentbox_system_prompt_overrides, do:true"`
	Code      any        // Prompt override code
	Content   any        // Prompt override content
	CreatedAt *time.Time // Creation time
	UpdatedAt *time.Time // Update time
}
