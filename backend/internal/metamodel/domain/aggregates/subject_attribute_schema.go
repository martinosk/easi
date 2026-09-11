package aggregates

import (
	"errors"

	"easi/backend/internal/metamodel/domain/events"
	"easi/backend/internal/metamodel/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

var (
	ErrAttributeNotFound       = errors.New("attribute not found on this subject type's schema")
	ErrAttributeRetired        = errors.New("attribute is retired")
	ErrAttributeAlreadyRetired = errors.New("attribute is already retired")
	ErrAttributeAlreadyActive  = errors.New("attribute is already active")
	ErrAttributeAlreadyDefined = errors.New("an attribute with this ID is already defined on this subject type's schema")
	ErrDuplicateAttributeName  = errors.New("an attribute with this name already exists on this subject type")
	ErrAttributeTypeImmutable  = errors.New("attribute types are immutable; retire this attribute and define a new one")
)

type SubjectAttributeSchema struct {
	domain.AggregateRoot
	tenantID    sharedvo.TenantID
	subjectType valueobjects.SubjectType
	attributes  []valueobjects.SubjectAttribute
	createdAt   valueobjects.Timestamp
	modifiedAt  valueobjects.Timestamp
	modifiedBy  valueobjects.UserEmail
}

func NewSubjectAttributeSchema(
	tenantID sharedvo.TenantID,
	subjectType valueobjects.SubjectType,
	createdBy valueobjects.UserEmail,
) (*SubjectAttributeSchema, error) {
	schema := &SubjectAttributeSchema{AggregateRoot: domain.NewAggregateRoot()}
	event := events.NewSubjectAttributeSchemaCreated(events.CreateSchemaParams{
		ID:          schema.ID(),
		TenantID:    tenantID.Value(),
		SubjectType: subjectType.Value(),
		CreatedBy:   createdBy.Value(),
	})
	if err := schema.applyAndRaise(event); err != nil {
		return nil, err
	}
	return schema, nil
}

func LoadSubjectAttributeSchemaFromHistory(eventHistory []domain.DomainEvent) (*SubjectAttributeSchema, error) {
	schema := &SubjectAttributeSchema{AggregateRoot: domain.NewAggregateRoot()}
	var applyErr error
	schema.LoadFromHistory(eventHistory, func(event domain.DomainEvent) {
		if applyErr != nil {
			return
		}
		applyErr = schema.apply(event)
	})
	if applyErr != nil {
		return nil, applyErr
	}
	return schema, nil
}

func (s *SubjectAttributeSchema) TenantID() sharedvo.TenantID {
	return s.tenantID
}

func (s *SubjectAttributeSchema) SubjectType() valueobjects.SubjectType {
	return s.subjectType
}

func (s *SubjectAttributeSchema) Attributes() []valueobjects.SubjectAttribute {
	attributes := make([]valueobjects.SubjectAttribute, len(s.attributes))
	copy(attributes, s.attributes)
	return attributes
}

func (s *SubjectAttributeSchema) CreatedAt() valueobjects.Timestamp {
	return s.createdAt
}

func (s *SubjectAttributeSchema) ModifiedAt() valueobjects.Timestamp {
	return s.modifiedAt
}

func (s *SubjectAttributeSchema) ModifiedBy() valueobjects.UserEmail {
	return s.modifiedBy
}

func (s *SubjectAttributeSchema) AttributeByID(id valueobjects.AttributeID) (valueobjects.SubjectAttribute, bool) {
	index := s.findAttributeIndex(id.Value())
	if index < 0 {
		return valueobjects.SubjectAttribute{}, false
	}
	return s.attributes[index], true
}

type DefineAttributeParams struct {
	Name         valueobjects.AttributeName
	Type         valueobjects.AttributeType
	HelpText     valueobjects.HelpText
	OptionLabels []valueobjects.OptionLabel
	Min          *float64
	Max          *float64
}

func (s *SubjectAttributeSchema) DefineAttribute(params DefineAttributeParams, modifiedBy valueobjects.UserEmail) (valueobjects.AttributeID, error) {
	options := make([]valueobjects.AttributeOption, len(params.OptionLabels))
	for i, label := range params.OptionLabels {
		options[i] = valueobjects.NewAttributeOption(valueobjects.NewOptionID(), label)
	}
	attribute, err := valueobjects.NewSubjectAttribute(valueobjects.SubjectAttributeParams{
		ID:       valueobjects.NewAttributeID(),
		Name:     params.Name,
		Type:     params.Type,
		HelpText: params.HelpText,
		Options:  options,
		Min:      params.Min,
		Max:      params.Max,
	})
	if err != nil {
		return valueobjects.AttributeID{}, err
	}
	if s.activeNameExists(params.Name, "") {
		return valueobjects.AttributeID{}, ErrDuplicateAttributeName
	}
	if err := s.applyAndRaise(s.definedEvent(attribute, modifiedBy)); err != nil {
		return valueobjects.AttributeID{}, err
	}
	return attribute.ID(), nil
}

func (s *SubjectAttributeSchema) ImportAttribute(attribute valueobjects.SubjectAttribute, modifiedBy valueobjects.UserEmail) error {
	if s.findAttributeIndex(attribute.ID().Value()) >= 0 {
		return ErrAttributeAlreadyDefined
	}
	if err := s.applyAndRaise(s.definedEvent(attribute, modifiedBy)); err != nil {
		return err
	}
	if attribute.IsActive() {
		return nil
	}
	return s.applyAndRaise(events.NewSubjectAttributeRetired(s.nextEventParams(attribute.ID(), modifiedBy)))
}

func (s *SubjectAttributeSchema) definedEvent(attribute valueobjects.SubjectAttribute, modifiedBy valueobjects.UserEmail) events.SubjectAttributeDefined {
	return events.NewSubjectAttributeDefined(s.nextEventParams(attribute.ID(), modifiedBy), events.AttributeDefinitionData{
		Name:          attribute.Name().Value(),
		AttributeType: attribute.Type().Value(),
		HelpText:      attribute.HelpText().Value(),
		Options:       optionsToEventData(attribute.Options()),
		Min:           attribute.Min(),
		Max:           attribute.Max(),
	})
}

type RenameAttributeParams struct {
	AttributeID   valueobjects.AttributeID
	Name          valueobjects.AttributeName
	HelpText      valueobjects.HelpText
	RequestedType string
}

func (s *SubjectAttributeSchema) RenameAttribute(params RenameAttributeParams, modifiedBy valueobjects.UserEmail) error {
	attribute, err := s.activeAttribute(params.AttributeID)
	if err != nil {
		return err
	}
	if params.RequestedType != "" && params.RequestedType != attribute.Type().Value() {
		return ErrAttributeTypeImmutable
	}
	if s.activeNameExists(params.Name, params.AttributeID.Value()) {
		return ErrDuplicateAttributeName
	}
	event := events.NewSubjectAttributeRenamed(s.nextEventParams(params.AttributeID, modifiedBy), params.Name.Value(), params.HelpText.Value())
	return s.applyAndRaise(event)
}

func (s *SubjectAttributeSchema) RetireAttribute(id valueobjects.AttributeID, modifiedBy valueobjects.UserEmail) error {
	return s.guardAndRaise(id, events.NewSubjectAttributeRetired(s.nextEventParams(id, modifiedBy)), func(attribute valueobjects.SubjectAttribute) error {
		if !attribute.IsActive() {
			return ErrAttributeAlreadyRetired
		}
		return nil
	})
}

func (s *SubjectAttributeSchema) ReactivateAttribute(id valueobjects.AttributeID, modifiedBy valueobjects.UserEmail) error {
	return s.guardAndRaise(id, events.NewSubjectAttributeReactivated(s.nextEventParams(id, modifiedBy)), func(attribute valueobjects.SubjectAttribute) error {
		if attribute.IsActive() {
			return ErrAttributeAlreadyActive
		}
		if s.activeNameExists(attribute.Name(), id.Value()) {
			return ErrDuplicateAttributeName
		}
		return nil
	})
}

func (s *SubjectAttributeSchema) AddOption(id valueobjects.AttributeID, label valueobjects.OptionLabel, modifiedBy valueobjects.UserEmail) (valueobjects.OptionID, error) {
	attribute, err := s.activeAttribute(id)
	if err != nil {
		return valueobjects.OptionID{}, err
	}
	optionID := valueobjects.NewOptionID()
	if _, err := attribute.WithAddedOption(valueobjects.NewAttributeOption(optionID, label)); err != nil {
		return valueobjects.OptionID{}, err
	}
	event := events.NewSubjectAttributeOptionAdded(s.nextEventParams(id, modifiedBy), optionID.Value(), label.Value())
	if err := s.applyAndRaise(event); err != nil {
		return valueobjects.OptionID{}, err
	}
	return optionID, nil
}

func (s *SubjectAttributeSchema) RetireOption(id valueobjects.AttributeID, optionID valueobjects.OptionID, modifiedBy valueobjects.UserEmail) error {
	event := events.NewSubjectAttributeOptionRetired(s.nextEventParams(id, modifiedBy), optionID.Value())
	return s.guardAndRaise(id, event, requireActiveThen(func(attribute valueobjects.SubjectAttribute) error {
		_, err := attribute.WithRetiredOption(optionID)
		return err
	}))
}

func (s *SubjectAttributeSchema) SetBounds(id valueobjects.AttributeID, min, max *float64, modifiedBy valueobjects.UserEmail) error {
	event := events.NewSubjectAttributeBoundsChanged(s.nextEventParams(id, modifiedBy), min, max)
	return s.guardAndRaise(id, event, requireActiveThen(func(attribute valueobjects.SubjectAttribute) error {
		_, err := attribute.WithBounds(min, max)
		return err
	}))
}

func requireActiveThen(next func(valueobjects.SubjectAttribute) error) func(valueobjects.SubjectAttribute) error {
	return func(attribute valueobjects.SubjectAttribute) error {
		if !attribute.IsActive() {
			return ErrAttributeRetired
		}
		return next(attribute)
	}
}

func (s *SubjectAttributeSchema) guardAndRaise(id valueobjects.AttributeID, event domain.DomainEvent, guard func(valueobjects.SubjectAttribute) error) error {
	attribute, err := s.attributeAt(id)
	if err != nil {
		return err
	}
	if err := guard(attribute); err != nil {
		return err
	}
	return s.applyAndRaise(event)
}

func (s *SubjectAttributeSchema) nextEventParams(id valueobjects.AttributeID, modifiedBy valueobjects.UserEmail) events.ModifySchemaParams {
	return events.ModifySchemaParams{
		SchemaID:    s.ID(),
		TenantID:    s.tenantID.Value(),
		SubjectType: s.subjectType.Value(),
		Version:     s.Version() + 1,
		AttributeID: id.Value(),
		ModifiedBy:  modifiedBy.Value(),
	}
}

func (s *SubjectAttributeSchema) activeAttribute(id valueobjects.AttributeID) (valueobjects.SubjectAttribute, error) {
	attribute, err := s.attributeAt(id)
	if err != nil {
		return valueobjects.SubjectAttribute{}, err
	}
	if !attribute.IsActive() {
		return valueobjects.SubjectAttribute{}, ErrAttributeRetired
	}
	return attribute, nil
}

func (s *SubjectAttributeSchema) attributeAt(id valueobjects.AttributeID) (valueobjects.SubjectAttribute, error) {
	index := s.findAttributeIndex(id.Value())
	if index < 0 {
		return valueobjects.SubjectAttribute{}, ErrAttributeNotFound
	}
	return s.attributes[index], nil
}

func (s *SubjectAttributeSchema) findAttributeIndex(id string) int {
	for i, attribute := range s.attributes {
		if attribute.ID().Value() == id {
			return i
		}
	}
	return -1
}

func (s *SubjectAttributeSchema) activeNameExists(name valueobjects.AttributeName, excludeID string) bool {
	for _, attribute := range s.attributes {
		if attribute.ID().Value() == excludeID || !attribute.IsActive() {
			continue
		}
		if attribute.Name().EqualsIgnoreCase(name) {
			return true
		}
	}
	return false
}

func optionsToEventData(options []valueobjects.AttributeOption) []events.AttributeOptionData {
	data := make([]events.AttributeOptionData, len(options))
	for i, option := range options {
		data[i] = events.AttributeOptionData{ID: option.ID().Value(), Label: option.Label().Value(), Active: option.IsActive()}
	}
	return data
}
