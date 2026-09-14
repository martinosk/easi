package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type ContainmentReference struct {
	ContainmentsID string `json:"containmentsId"`
	PartID         string `json:"partId"`
	ParentID       string `json:"parentId"`
}

func (r ContainmentReference) aggregateID(baseID string) string {
	if baseID != "" {
		return baseID
	}
	return r.ContainmentsID
}

type ComponentAttached struct {
	domain.BaseEvent
	ContainmentReference
	Kind       string    `json:"kind"`
	AttachedAt time.Time `json:"attachedAt"`
}

func NewComponentAttached(containmentsID, partID, parentID, kind string) ComponentAttached {
	return ComponentAttached{
		BaseEvent:            domain.NewBaseEvent(containmentsID),
		ContainmentReference: ContainmentReference{ContainmentsID: containmentsID, PartID: partID, ParentID: parentID},
		Kind:                 kind,
		AttachedAt:           time.Now().UTC(),
	}
}

func (e ComponentAttached) AggregateID() string {
	return e.aggregateID(e.BaseEvent.AggregateID())
}

func (e ComponentAttached) EventType() string {
	return "ComponentAttached"
}

func (e ComponentAttached) EventData() map[string]any {
	return map[string]any{
		"containmentsId": e.ContainmentsID,
		"partId":         e.PartID,
		"parentId":       e.ParentID,
		"kind":           e.Kind,
		"attachedAt":     e.AttachedAt,
	}
}

type ComponentDetached struct {
	domain.BaseEvent
	ContainmentReference
	DetachedAt time.Time `json:"detachedAt"`
}

func NewComponentDetached(containmentsID, partID, parentID string) ComponentDetached {
	return ComponentDetached{
		BaseEvent:            domain.NewBaseEvent(containmentsID),
		ContainmentReference: ContainmentReference{ContainmentsID: containmentsID, PartID: partID, ParentID: parentID},
		DetachedAt:           time.Now().UTC(),
	}
}

func (e ComponentDetached) AggregateID() string {
	return e.aggregateID(e.BaseEvent.AggregateID())
}

func (e ComponentDetached) EventType() string {
	return "ComponentDetached"
}

func (e ComponentDetached) EventData() map[string]any {
	return map[string]any{
		"containmentsId": e.ContainmentsID,
		"partId":         e.PartID,
		"parentId":       e.ParentID,
		"detachedAt":     e.DetachedAt,
	}
}
