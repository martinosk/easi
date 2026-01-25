package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type ComponentRelationCreated struct {
	domain.BaseEvent
	ID                string    `json:"id"`
	SourceComponentID string    `json:"sourceComponentId"`
	TargetComponentID string    `json:"targetComponentId"`
	RelationType      string    `json:"relationType"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	CreatedAt         time.Time `json:"createdAt"`
}

func (e ComponentRelationCreated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

type ComponentRelationParams struct {
	ID          string
	SourceID    string
	TargetID    string
	Type        string
	Name        string
	Description string
}

func NewComponentRelationCreated(params ComponentRelationParams) ComponentRelationCreated {
	return ComponentRelationCreated{
		BaseEvent:         domain.NewBaseEvent(params.ID),
		ID:                params.ID,
		SourceComponentID: params.SourceID,
		TargetComponentID: params.TargetID,
		RelationType:      params.Type,
		Name:              params.Name,
		Description:       params.Description,
		CreatedAt:         time.Now().UTC(),
	}
}

func (e ComponentRelationCreated) EventType() string {
	return "ComponentRelationCreated"
}

func (e ComponentRelationCreated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":                e.ID,
		"sourceComponentId": e.SourceComponentID,
		"targetComponentId": e.TargetComponentID,
		"relationType":      e.RelationType,
		"name":              e.Name,
		"description":       e.Description,
		"createdAt":         e.CreatedAt,
	}
}
