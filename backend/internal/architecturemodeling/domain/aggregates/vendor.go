package aggregates

import (
	"time"

	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
)

type Vendor struct {
	domain.AggregateRoot
	name                  valueobjects.EntityName
	implementationPartner string
	notes                 valueobjects.Notes
	createdAt             time.Time
	isDeleted             bool
}

func NewVendor(
	name valueobjects.EntityName,
	implementationPartner string,
	notes valueobjects.Notes,
) (*Vendor, error) {
	aggregate := &Vendor{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	event := events.NewVendorCreated(
		aggregate.ID(),
		name.Value(),
		implementationPartner,
		notes.Value(),
	)

	aggregate.apply(event)
	aggregate.RaiseEvent(event)

	return aggregate, nil
}

func LoadVendorFromHistory(events []domain.DomainEvent) (*Vendor, error) {
	aggregate := &Vendor{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	aggregate.LoadFromHistory(events, func(event domain.DomainEvent) {
		aggregate.apply(event)
	})

	return aggregate, nil
}

func (v *Vendor) Update(
	name valueobjects.EntityName,
	implementationPartner string,
	notes valueobjects.Notes,
) error {
	event := events.NewVendorUpdated(
		v.ID(),
		name.Value(),
		implementationPartner,
		notes.Value(),
	)

	v.apply(event)
	v.RaiseEvent(event)

	return nil
}

func (v *Vendor) Delete() error {
	event := events.NewVendorDeleted(
		v.ID(),
		v.name.Value(),
	)

	v.apply(event)
	v.RaiseEvent(event)

	return nil
}

func (v *Vendor) apply(event domain.DomainEvent) {
	switch e := event.(type) {
	case events.VendorCreated:
		v.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
		v.name = valueobjects.MustNewEntityName(e.Name)
		v.implementationPartner = e.ImplementationPartner
		v.notes = valueobjects.MustNewNotes(e.Notes)
		v.createdAt = e.CreatedAt
	case events.VendorUpdated:
		v.name = valueobjects.MustNewEntityName(e.Name)
		v.implementationPartner = e.ImplementationPartner
		v.notes = valueobjects.MustNewNotes(e.Notes)
	case events.VendorDeleted:
		v.isDeleted = true
	}
}

func (v *Vendor) Name() valueobjects.EntityName {
	return v.name
}

func (v *Vendor) ImplementationPartner() string {
	return v.implementationPartner
}

func (v *Vendor) Notes() valueobjects.Notes {
	return v.notes
}

func (v *Vendor) CreatedAt() time.Time {
	return v.createdAt
}

func (v *Vendor) IsDeleted() bool {
	return v.isDeleted
}
