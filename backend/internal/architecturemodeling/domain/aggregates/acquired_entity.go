package aggregates

import (
	"time"

	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
)

type AcquiredEntity struct {
	domain.AggregateRoot
	name              valueobjects.EntityName
	acquisitionDate   *time.Time
	integrationStatus valueobjects.IntegrationStatus
	notes             valueobjects.Notes
	createdAt         time.Time
	isDeleted         bool
}

func NewAcquiredEntity(
	name valueobjects.EntityName,
	acquisitionDate *time.Time,
	integrationStatus valueobjects.IntegrationStatus,
	notes valueobjects.Notes,
) (*AcquiredEntity, error) {
	aggregate := &AcquiredEntity{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	event := events.NewAcquiredEntityCreated(
		aggregate.ID(),
		name.Value(),
		acquisitionDate,
		integrationStatus.Value(),
		notes.Value(),
	)

	aggregate.apply(event)
	aggregate.RaiseEvent(event)

	return aggregate, nil
}

func LoadAcquiredEntityFromHistory(events []domain.DomainEvent) (*AcquiredEntity, error) {
	aggregate := &AcquiredEntity{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	aggregate.LoadFromHistory(events, func(event domain.DomainEvent) {
		aggregate.apply(event)
	})

	return aggregate, nil
}

func (a *AcquiredEntity) Update(
	name valueobjects.EntityName,
	acquisitionDate *time.Time,
	integrationStatus valueobjects.IntegrationStatus,
	notes valueobjects.Notes,
) error {
	event := events.NewAcquiredEntityUpdated(
		a.ID(),
		name.Value(),
		acquisitionDate,
		integrationStatus.Value(),
		notes.Value(),
	)

	a.apply(event)
	a.RaiseEvent(event)

	return nil
}

func (a *AcquiredEntity) Delete() error {
	event := events.NewAcquiredEntityDeleted(
		a.ID(),
		a.name.Value(),
	)

	a.apply(event)
	a.RaiseEvent(event)

	return nil
}

func (a *AcquiredEntity) apply(event domain.DomainEvent) {
	switch e := event.(type) {
	case events.AcquiredEntityCreated:
		a.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
		a.name = valueobjects.MustNewEntityName(e.Name)
		a.acquisitionDate = e.AcquisitionDate
		a.integrationStatus = valueobjects.MustNewIntegrationStatus(e.IntegrationStatus)
		a.notes = valueobjects.MustNewNotes(e.Notes)
		a.createdAt = e.CreatedAt
	case events.AcquiredEntityUpdated:
		a.name = valueobjects.MustNewEntityName(e.Name)
		a.acquisitionDate = e.AcquisitionDate
		a.integrationStatus = valueobjects.MustNewIntegrationStatus(e.IntegrationStatus)
		a.notes = valueobjects.MustNewNotes(e.Notes)
	case events.AcquiredEntityDeleted:
		a.isDeleted = true
	}
}

func (a *AcquiredEntity) Name() valueobjects.EntityName {
	return a.name
}

func (a *AcquiredEntity) AcquisitionDate() *time.Time {
	return a.acquisitionDate
}

func (a *AcquiredEntity) IntegrationStatus() valueobjects.IntegrationStatus {
	return a.integrationStatus
}

func (a *AcquiredEntity) Notes() valueobjects.Notes {
	return a.notes
}

func (a *AcquiredEntity) CreatedAt() time.Time {
	return a.createdAt
}

func (a *AcquiredEntity) IsDeleted() bool {
	return a.isDeleted
}
