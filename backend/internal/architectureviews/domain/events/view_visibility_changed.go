package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
)

type ViewVisibilityChanged struct {
	domain.BaseEvent
	ViewID      string `json:"viewId"`
	IsPrivate   bool   `json:"isPrivate"`
	OwnerUserID string `json:"ownerUserId"`
	OwnerEmail  string `json:"ownerEmail"`
}

func (e ViewVisibilityChanged) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ViewID
}

func NewViewVisibilityChanged(viewID string, isPrivate bool, ownerUserID, ownerEmail string) ViewVisibilityChanged {
	return ViewVisibilityChanged{
		BaseEvent:   domain.NewBaseEvent(viewID),
		ViewID:      viewID,
		IsPrivate:   isPrivate,
		OwnerUserID: ownerUserID,
		OwnerEmail:  ownerEmail,
	}
}

func (e ViewVisibilityChanged) EventType() string {
	return "ViewVisibilityChanged"
}

func (e ViewVisibilityChanged) EventData() map[string]interface{} {
	return map[string]interface{}{
		"viewId":      e.ViewID,
		"isPrivate":   e.IsPrivate,
		"ownerUserId": e.OwnerUserID,
		"ownerEmail":  e.OwnerEmail,
	}
}
