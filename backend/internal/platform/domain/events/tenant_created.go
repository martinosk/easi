package events

import (
	"easi/backend/internal/shared/eventsourcing"
	"time"
)

type TenantCreated struct {
	domain.BaseEvent
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Status          string    `json:"status"`
	Domains         []string  `json:"domains"`
	FirstAdminEmail string    `json:"firstAdminEmail"`
	CreatedAt       time.Time `json:"createdAt"`
}

func NewTenantCreated(
	id, name, status string,
	domains []string,
	firstAdminEmail string,
) TenantCreated {
	return TenantCreated{
		BaseEvent:       domain.NewBaseEvent(id),
		ID:              id,
		Name:            name,
		Status:          status,
		Domains:         domains,
		FirstAdminEmail: firstAdminEmail,
		CreatedAt:       time.Now().UTC(),
	}
}

func (e TenantCreated) EventType() string {
	return "TenantCreated"
}

func (e TenantCreated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":              e.ID,
		"name":            e.Name,
		"status":          e.Status,
		"domains":         e.Domains,
		"firstAdminEmail": e.FirstAdminEmail,
		"createdAt":       e.CreatedAt,
	}
}
