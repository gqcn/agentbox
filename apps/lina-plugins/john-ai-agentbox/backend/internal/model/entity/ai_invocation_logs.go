// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// AiInvocationLogs is the golang structure for table ai_invocation_logs.
type AiInvocationLogs struct {
	Id              int64      `json:"id"              orm:"id"                description:"AI invocation log ID"`
	Purpose         string     `json:"purpose"         orm:"purpose"           description:"Invocation purpose"`
	TierCode        string     `json:"tierCode"        orm:"tier_code"         description:"Capability tier code"`
	ProviderId      int64      `json:"providerId"      orm:"provider_id"       description:"Provider ID"`
	ProviderModelId int64      `json:"providerModelId" orm:"provider_model_id" description:"Provider model ID"`
	ModelName       string     `json:"modelName"       orm:"model_name"        description:"Model name used for the invocation"`
	Protocol        string     `json:"protocol"        orm:"protocol"          description:"Model protocol"`
	Status          string     `json:"status"          orm:"status"            description:"Invocation status"`
	LatencyMs       int64      `json:"latencyMs"       orm:"latency_ms"        description:"Invocation latency in milliseconds"`
	ErrorMessage    string     `json:"errorMessage"    orm:"error_message"     description:"Invocation error message"`
	CreatedAt       *time.Time `json:"createdAt"       orm:"created_at"        description:"Creation time"`
}
