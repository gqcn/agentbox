// This file holds coding-agent query projections. The service uses one joined
// projection for list and detail so provider, image, and runtime fields are
// loaded in one database query per request.

package catalog

import "time"

// agentJoinRow matches the aliased projection returned by agentSelectFields.
type agentJoinRow struct {
	ID            string     `orm:"id"`
	UserID        string     `orm:"user_id"`
	Name          string     `orm:"name"`
	ProviderID    int64      `orm:"provider_id"`
	ProviderName  string     `orm:"provider_name"`
	ModelName     string     `orm:"model_name"`
	ModelProtocol string     `orm:"model_protocol"`
	ImageID       int64      `orm:"image_id"`
	ImageName     string     `orm:"image_name"`
	ImageRef      string     `orm:"image_ref"`
	AgentType     string     `orm:"agent_type"`
	IconKey       string     `orm:"icon_key"`
	Notes         string     `orm:"notes"`
	RuntimeStatus string     `orm:"runtime_status"`
	ContainerID   string     `orm:"container_id"`
	DockerID      string     `orm:"docker_id"`
	DeletedAt     *time.Time `orm:"deleted_at"`
	CreatedAt     *time.Time `orm:"created_at"`
	UpdatedAt     *time.Time `orm:"updated_at"`
}

// agentSelectFields returns the stable projection needed by agent list/detail APIs.
func agentSelectFields() string {
	return `
		a.id,
		a.user_id,
		a.name,
		a.provider_id,
		p.name AS provider_name,
		a.model_name,
		a.model_protocol,
		a.image_id,
		i.name AS image_name,
		i.image_ref,
		a.agent_type,
		a.icon_key,
		a.notes,
		COALESCE(r.status, '') AS runtime_status,
		COALESCE(r.container_id, '') AS container_id,
		COALESCE(r.docker_id, '') AS docker_id,
		a.deleted_at,
		a.created_at,
		a.updated_at
	`
}

// agentInfoFromJoinRow maps one joined projection to the service API shape.
func agentInfoFromJoinRow(row *agentJoinRow) AgentInfo {
	if row == nil {
		return AgentInfo{}
	}
	item := AgentInfo{
		ID:             row.ID,
		UserID:         row.UserID,
		Name:           row.Name,
		ProviderID:     row.ProviderID,
		ProviderName:   row.ProviderName,
		ModelName:      row.ModelName,
		ModelProtocol:  row.ModelProtocol,
		ImageID:        row.ImageID,
		ImageName:      row.ImageName,
		ImageRef:       row.ImageRef,
		AgentType:      row.AgentType,
		IconKey:        row.IconKey,
		Notes:          row.Notes,
		RuntimeStatus:  row.RuntimeStatus,
		ActivityStatus: row.RuntimeStatus,
		ContainerID:    row.ContainerID,
		DockerID:       row.DockerID,
		CreatedAt:      unixMilliFromTimePtr(row.CreatedAt),
		UpdatedAt:      unixMilliFromTimePtr(row.UpdatedAt),
	}
	if value := unixMilliFromTimePtr(row.DeletedAt); value > 0 {
		item.DeletedAt = &value
	}
	return item
}
