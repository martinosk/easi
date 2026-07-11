package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/onepagers/application/commands"
	"easi/backend/internal/onepagers/application/ports"
	"easi/backend/internal/onepagers/application/readmodels"
	"easi/backend/internal/onepagers/domain/aggregates"
	"easi/backend/internal/onepagers/domain/valueobjects"
	"easi/backend/internal/onepagers/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

var (
	ErrFieldNotDefined        = errors.New("field is not defined on the subject type's one-pager configuration")
	ErrFieldDefinitionRetired = errors.New("field is retired on the subject type's one-pager configuration")
	ErrValueTypeMismatch      = errors.New("value type does not match the field's type")
	ErrOptionNotDefined       = errors.New("selection option is not defined on this field")
	ErrOptionRetired          = errors.New("selection option is retired")
	ErrSubjectNotFound        = errors.New("subject does not exist")
)

type ConfigurationDefinitions interface {
	GetBySubjectType(ctx context.Context, subjectType string) (*readmodels.ConfigurationRecord, error)
}

type FactsLookup interface {
	FactsIDForSubject(ctx context.Context, subject readmodels.SubjectKey) (string, error)
}

func subjectKeyFor(subjectRef valueobjects.SubjectRef) readmodels.SubjectKey {
	return readmodels.SubjectKey{
		SubjectType: subjectRef.SubjectType().Value(),
		SubjectID:   subjectRef.SubjectID(),
	}
}

type factsWriteInput struct {
	tenantID   sharedvo.TenantID
	subjectRef valueobjects.SubjectRef
	fieldID    valueobjects.FieldID
	modifiedBy valueobjects.UserEmail
}

type rawFactsWrite struct {
	tenantID    string
	subjectType string
	subjectID   string
	fieldID     string
	modifiedBy  string
}

func parseFactsWriteInput(raw rawFactsWrite) (factsWriteInput, error) {
	parsedTenant, err := sharedvo.NewTenantID(raw.tenantID)
	if err != nil {
		return factsWriteInput{}, err
	}
	subjectRef, err := valueobjects.NewSubjectRef(raw.subjectType, raw.subjectID)
	if err != nil {
		return factsWriteInput{}, err
	}
	parsedFieldID, err := valueobjects.NewFieldIDFromString(raw.fieldID)
	if err != nil {
		return factsWriteInput{}, err
	}
	email, err := valueobjects.NewUserEmail(raw.modifiedBy)
	if err != nil {
		return factsWriteInput{}, err
	}
	return factsWriteInput{tenantID: parsedTenant, subjectRef: subjectRef, fieldID: parsedFieldID, modifiedBy: email}, nil
}

func activeFieldDefinition(
	ctx context.Context,
	configs ConfigurationDefinitions,
	subjectType, fieldID string,
) (readmodels.CustomFieldRecord, error) {
	record, err := configs.GetBySubjectType(ctx, subjectType)
	if err != nil {
		return readmodels.CustomFieldRecord{}, err
	}
	if record == nil {
		return readmodels.CustomFieldRecord{}, ErrFieldNotDefined
	}
	for _, field := range record.Document.CustomFields {
		if field.ID != fieldID {
			continue
		}
		if !field.Active {
			return readmodels.CustomFieldRecord{}, ErrFieldDefinitionRetired
		}
		return field, nil
	}
	return readmodels.CustomFieldRecord{}, ErrFieldNotDefined
}

func validateSelectionOption(field readmodels.CustomFieldRecord, value valueobjects.FieldValue) error {
	selection, ok := value.(valueobjects.SelectionValue)
	if !ok {
		return nil
	}
	for _, option := range field.Options {
		if option.ID != selection.OptionID().Value() {
			continue
		}
		if !option.Active {
			return ErrOptionRetired
		}
		return nil
	}
	return ErrOptionNotDefined
}

type RecordFieldValueHandler struct {
	repository *repositories.OnePagerFactsRepository
	configs    ConfigurationDefinitions
	facts      FactsLookup
	subjects   ports.SubjectExistenceChecker
}

func NewRecordFieldValueHandler(
	repository *repositories.OnePagerFactsRepository,
	configs ConfigurationDefinitions,
	facts FactsLookup,
	subjects ports.SubjectExistenceChecker,
) *RecordFieldValueHandler {
	return &RecordFieldValueHandler{repository: repository, configs: configs, facts: facts, subjects: subjects}
}

func (h *RecordFieldValueHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.RecordFieldValue)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	input, err := parseFactsWriteInput(rawFactsWrite{
		tenantID:    command.TenantID,
		subjectType: command.SubjectType,
		subjectID:   command.SubjectID,
		fieldID:     command.FieldID,
		modifiedBy:  command.ModifiedBy,
	})
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	value, err := h.validateValue(ctx, command, input)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	facts, err := h.loadOrCreateFacts(ctx, input)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := facts.RecordFieldValue(input.fieldID, value, input.modifiedBy); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, facts); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(facts.ID()), nil
}

func (h *RecordFieldValueHandler) validateValue(
	ctx context.Context,
	command *commands.RecordFieldValue,
	input factsWriteInput,
) (valueobjects.FieldValue, error) {
	field, err := activeFieldDefinition(ctx, h.configs, input.subjectRef.SubjectType().Value(), input.fieldID.Value())
	if err != nil {
		return nil, err
	}
	if command.Value.Type != field.Type {
		return nil, ErrValueTypeMismatch
	}
	value, err := valueobjects.FieldValueFromEnvelope(command.Value)
	if err != nil {
		return nil, err
	}
	if err := validateSelectionOption(field, value); err != nil {
		return nil, err
	}
	return value, nil
}

func (h *RecordFieldValueHandler) loadOrCreateFacts(ctx context.Context, input factsWriteInput) (*aggregates.OnePagerFacts, error) {
	factsID, err := h.facts.FactsIDForSubject(ctx, subjectKeyFor(input.subjectRef))
	if err != nil {
		return nil, err
	}
	if factsID != "" {
		return h.repository.GetByID(ctx, factsID)
	}

	exists, err := h.subjects.SubjectExists(ctx, input.subjectRef.SubjectType().Value(), input.subjectRef.SubjectID())
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrSubjectNotFound
	}
	return aggregates.NewOnePagerFacts(input.tenantID, input.subjectRef), nil
}

type ClearFieldValueHandler struct {
	repository *repositories.OnePagerFactsRepository
	configs    ConfigurationDefinitions
	facts      FactsLookup
}

func NewClearFieldValueHandler(
	repository *repositories.OnePagerFactsRepository,
	configs ConfigurationDefinitions,
	facts FactsLookup,
) *ClearFieldValueHandler {
	return &ClearFieldValueHandler{repository: repository, configs: configs, facts: facts}
}

func (h *ClearFieldValueHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.ClearFieldValue)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	input, err := parseFactsWriteInput(rawFactsWrite{
		tenantID:    command.TenantID,
		subjectType: command.SubjectType,
		subjectID:   command.SubjectID,
		fieldID:     command.FieldID,
		modifiedBy:  command.ModifiedBy,
	})
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if _, err := activeFieldDefinition(ctx, h.configs, input.subjectRef.SubjectType().Value(), input.fieldID.Value()); err != nil {
		return cqrs.EmptyResult(), err
	}

	return h.clearExistingValue(ctx, input)
}

func (h *ClearFieldValueHandler) clearExistingValue(ctx context.Context, input factsWriteInput) (cqrs.CommandResult, error) {
	factsID, err := h.facts.FactsIDForSubject(ctx, subjectKeyFor(input.subjectRef))
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if factsID == "" {
		return cqrs.EmptyResult(), nil
	}

	facts, err := h.repository.GetByID(ctx, factsID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := facts.ClearFieldValue(input.fieldID, input.modifiedBy); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, facts); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(facts.ID()), nil
}

type ArchiveOnePagerFactsHandler struct {
	repository *repositories.OnePagerFactsRepository
}

func NewArchiveOnePagerFactsHandler(repository *repositories.OnePagerFactsRepository) *ArchiveOnePagerFactsHandler {
	return &ArchiveOnePagerFactsHandler{repository: repository}
}

func (h *ArchiveOnePagerFactsHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.ArchiveOnePagerFacts)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	facts, err := h.repository.GetByID(ctx, command.FactsID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := facts.Archive(command.Reason); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, facts); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(facts.ID()), nil
}
