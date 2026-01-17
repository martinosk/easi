package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type PurchasedFromRelationshipDeleted struct {
	domain.BaseEvent
	ID          string
	VendorID    string
	ComponentID string
	DeletedAt   time.Time
}

func NewPurchasedFromRelationshipDeleted(id, vendorID, componentID string) PurchasedFromRelationshipDeleted {
	return PurchasedFromRelationshipDeleted{
		BaseEvent:   domain.NewBaseEvent(id),
		ID:          id,
		VendorID:    vendorID,
		ComponentID: componentID,
		DeletedAt:   time.Now().UTC(),
	}
}

func (e PurchasedFromRelationshipDeleted) EventType() string {
	return "PurchasedFromRelationshipDeleted"
}

func (e PurchasedFromRelationshipDeleted) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":          e.ID,
		"vendorId":    e.VendorID,
		"componentId": e.ComponentID,
		"deletedAt":   e.DeletedAt,
	}
}
