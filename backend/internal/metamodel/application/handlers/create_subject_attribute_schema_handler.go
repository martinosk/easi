package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/metamodel/application/commands"
	"easi/backend/internal/metamodel/application/readmodels"
	"easi/backend/internal/metamodel/domain/aggregates"
	"easi/backend/internal/metamodel/domain/valueobjects"
	"easi/backend/internal/metamodel/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

var ErrSchemaAlreadyExists = errors.New("a subject attribute schema already exists for this subject type")

type SchemaLookup interface {
	GetBySubjectType(ctx context.Context, subjectType string) (*readmodels.SubjectAttributeSchemaRecord, error)
}

type CreateSubjectAttributeSchemaHandler struct {
	repository *repositories.SubjectAttributeSchemaRepository
	lookup     SchemaLookup
}

func NewCreateSubjectAttributeSchemaHandler(repository *repositories.SubjectAttributeSchemaRepository, lookup SchemaLookup) *CreateSubjectAttributeSchemaHandler {
	return &CreateSubjectAttributeSchemaHandler{repository: repository, lookup: lookup}
}

func (h *CreateSubjectAttributeSchemaHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.CreateSubjectAttributeSchema)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}
	existing, err := h.lookup.GetBySubjectType(ctx, command.SubjectType)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if existing != nil {
		return cqrs.EmptyResult(), ErrSchemaAlreadyExists
	}
	schema, err := newSchema(command.TenantID, command.SubjectType, command.CreatedBy)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := h.repository.Save(ctx, schema); err != nil {
		return cqrs.EmptyResult(), err
	}
	return cqrs.NewResult(schema.ID()), nil
}

func newSchema(rawTenantID, rawSubjectType, rawCreatedBy string) (*aggregates.SubjectAttributeSchema, error) {
	tenantID, err := sharedvo.NewTenantID(rawTenantID)
	if err != nil {
		return nil, err
	}
	subjectType, err := valueobjects.NewSubjectType(rawSubjectType)
	if err != nil {
		return nil, err
	}
	createdBy, err := valueobjects.NewUserEmail(rawCreatedBy)
	if err != nil {
		return nil, err
	}
	return aggregates.NewSubjectAttributeSchema(tenantID, subjectType, createdBy)
}
