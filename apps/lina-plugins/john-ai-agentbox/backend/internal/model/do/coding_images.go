// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// CodingImages is the golang structure of table john_ai_agentbox_coding_images for DAO operations like Where/Data.
type CodingImages struct {
	g.Meta       `orm:"table:john_ai_agentbox_coding_images, do:true"`
	Id           any        // Coding image ID
	Name         any        // Image display name
	ImageRef     any        // Container image reference
	AgentType    any        // Agent runtime type
	DefaultShell any        // Default shell path
	Notes        any        // Image notes
	Enabled      any        // Whether the image is enabled
	IsDefault    any        // Whether the image is a default option
	CreatedAt    *time.Time // Creation time
	UpdatedAt    *time.Time // Update time
}
