package aggregates

import (
	"errors"

	"easi/backend/internal/onepagers/domain/events"
	"easi/backend/internal/onepagers/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

const ArchiveReasonSubjectDeleted = "subject deleted"

var (
	ErrFieldIDRequired    = errors.New("a field value requires a field ID")
	ErrFieldValueRequired = errors.New("a field value is required")
	ErrFactsArchived      = errors.New("one-pager facts are archived and can no longer be modified")
)

type OnePagerFacts struct {
	domain.AggregateRoot
	tenantID   sharedvo.TenantID
	subjectRef valueobjects.SubjectRef
	values     map[string]valueobjects.FieldValue
	archived   bool
}

func NewOnePagerFacts(tenantID sharedvo.TenantID, subjectRef valueobjects.SubjectRef) *OnePagerFacts {
	return &OnePagerFacts{
		AggregateRoot: domain.NewAggregateRoot(),
		tenantID:      tenantID,
		subjectRef:    subjectRef,
		values:        make(map[string]valueobjects.FieldValue),
	}
}

func LoadOnePagerFactsFromHistory(eventHistory []domain.DomainEvent) (*OnePagerFacts, error) {
	facts := &OnePagerFacts{
		AggregateRoot: domain.NewAggregateRoot(),
		values:        make(map[string]valueobjects.FieldValue),
	}

	var applyErr error
	facts.LoadFromHistory(eventHistory, func(event domain.DomainEvent) {
		if applyErr != nil {
			return
		}
		applyErr = facts.apply(event)
	})
	if applyErr != nil {
		return nil, applyErr
	}

	return facts, nil
}

func (f *OnePagerFacts) TenantID() sharedvo.TenantID {
	return f.tenantID
}

func (f *OnePagerFacts) SubjectRef() valueobjects.SubjectRef {
	return f.subjectRef
}

func (f *OnePagerFacts) IsArchived() bool {
	return f.archived
}

func (f *OnePagerFacts) Values() map[string]valueobjects.FieldValue {
	values := make(map[string]valueobjects.FieldValue, len(f.values))
	for fieldID, value := range f.values {
		values[fieldID] = value
	}
	return values
}

func (f *OnePagerFacts) Value(fieldID valueobjects.FieldID) (valueobjects.FieldValue, bool) {
	value, found := f.values[fieldID.Value()]
	return value, found
}

func (f *OnePagerFacts) RecordFieldValue(
	fieldID valueobjects.FieldID,
	value valueobjects.FieldValue,
	recordedBy valueobjects.UserEmail,
) error {
	if err := f.guardWrite(fieldID); err != nil {
		return err
	}
	if value == nil {
		return ErrFieldValueRequired
	}
	if current, found := f.values[fieldID.Value()]; found && current.Equals(value) {
		return nil
	}

	envelope, err := valueobjects.NewValueEnvelope(value)
	if err != nil {
		return err
	}

	event := events.NewFieldValueRecorded(f.nextEventParams(recordedBy.Value()), fieldID.Value(), envelope)
	return f.applyAndRaise(event)
}

func (f *OnePagerFacts) ClearFieldValue(fieldID valueobjects.FieldID, clearedBy valueobjects.UserEmail) error {
	if err := f.guardWrite(fieldID); err != nil {
		return err
	}
	if _, found := f.values[fieldID.Value()]; !found {
		return nil
	}

	event := events.NewFieldValueCleared(f.nextEventParams(clearedBy.Value()), fieldID.Value())
	return f.applyAndRaise(event)
}

func (f *OnePagerFacts) Archive(reason string) error {
	if f.archived {
		return nil
	}

	event := events.NewOnePagerFactsArchived(f.nextEventParams("system"), reason)
	return f.applyAndRaise(event)
}

func (f *OnePagerFacts) guardWrite(fieldID valueobjects.FieldID) error {
	if f.archived {
		return ErrFactsArchived
	}
	if fieldID.Value() == "" {
		return ErrFieldIDRequired
	}
	return nil
}

func (f *OnePagerFacts) nextEventParams(modifiedBy string) events.ModifyFactsParams {
	return events.ModifyFactsParams{
		FactsID:     f.ID(),
		TenantID:    f.tenantID.Value(),
		SubjectType: f.subjectRef.SubjectType().Value(),
		SubjectID:   f.subjectRef.SubjectID(),
		Version:     f.Version() + 1,
		ModifiedBy:  modifiedBy,
	}
}
