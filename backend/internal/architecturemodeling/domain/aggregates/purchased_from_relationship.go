package aggregates

import (
	"time"

	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
)

type PurchasedFromRelationship struct {
	domain.AggregateRoot
	vendorID    valueobjects.VendorID
	componentID valueobjects.ComponentID
	notes       valueobjects.Notes
	createdAt   time.Time
	isDeleted   bool
}

func NewPurchasedFromRelationship(
	vendorID valueobjects.VendorID,
	componentID valueobjects.ComponentID,
	notes valueobjects.Notes,
) (*PurchasedFromRelationship, error) {
	aggregate := &PurchasedFromRelationship{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	event := events.NewPurchasedFromRelationshipCreated(
		aggregate.ID(),
		vendorID.Value(),
		componentID.Value(),
		notes.Value(),
	)

	aggregate.apply(event)
	aggregate.RaiseEvent(event)

	return aggregate, nil
}

func LoadPurchasedFromRelationshipFromHistory(events []domain.DomainEvent) (*PurchasedFromRelationship, error) {
	aggregate := &PurchasedFromRelationship{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	aggregate.LoadFromHistory(events, func(event domain.DomainEvent) {
		aggregate.apply(event)
	})

	return aggregate, nil
}

func (p *PurchasedFromRelationship) Delete() error {
	event := events.NewPurchasedFromRelationshipDeleted(
		p.ID(),
		p.vendorID.Value(),
		p.componentID.Value(),
	)

	p.apply(event)
	p.RaiseEvent(event)

	return nil
}

func (p *PurchasedFromRelationship) apply(event domain.DomainEvent) {
	switch e := event.(type) {
	case events.PurchasedFromRelationshipCreated:
		p.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
		p.vendorID, _ = valueobjects.NewVendorIDFromString(e.VendorID)
		p.componentID, _ = valueobjects.NewComponentIDFromString(e.ComponentID)
		p.notes = valueobjects.MustNewNotes(e.Notes)
		p.createdAt = e.CreatedAt
	case events.PurchasedFromRelationshipDeleted:
		p.isDeleted = true
	}
}

func (p *PurchasedFromRelationship) VendorID() valueobjects.VendorID {
	return p.vendorID
}

func (p *PurchasedFromRelationship) ComponentID() valueobjects.ComponentID {
	return p.componentID
}

func (p *PurchasedFromRelationship) Notes() valueobjects.Notes {
	return p.notes
}

func (p *PurchasedFromRelationship) CreatedAt() time.Time {
	return p.createdAt
}

func (p *PurchasedFromRelationship) IsDeleted() bool {
	return p.isDeleted
}
