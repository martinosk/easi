package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type PurchasedFromRelationshipCreated struct {
	domain.BaseEvent
	ID          string
	VendorID    string
	ComponentID string
	Notes       string
	CreatedAt   time.Time
}

func NewPurchasedFromRelationshipCreated(id, vendorID, componentID, notes string) PurchasedFromRelationshipCreated {
	return PurchasedFromRelationshipCreated{
		BaseEvent:   domain.NewBaseEvent(id),
		ID:          id,
		VendorID:    vendorID,
		ComponentID: componentID,
		Notes:       notes,
		CreatedAt:   time.Now().UTC(),
	}
}

func (e PurchasedFromRelationshipCreated) EventType() string {
	return "PurchasedFromRelationshipCreated"
}

func (e PurchasedFromRelationshipCreated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":          e.ID,
		"vendorId":    e.VendorID,
		"componentId": e.ComponentID,
		"notes":       e.Notes,
		"createdAt":   e.CreatedAt,
	}
}
