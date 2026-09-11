package handlers

import (
	"context"

	"easi/backend/internal/metamodel/domain/aggregates"
	"easi/backend/internal/metamodel/domain/valueobjects"
	"easi/backend/internal/metamodel/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type schemaCommand interface {
	cqrs.Command
	SubjectSchemaID() string
	ModifiedByEmail() string
}

type schemaAction[C schemaCommand] func(schema *aggregates.SubjectAttributeSchema, cmd C, modifiedBy valueobjects.UserEmail) (string, error)

type modifySchemaHandler[C schemaCommand] struct {
	repository *repositories.SubjectAttributeSchemaRepository
	act        schemaAction[C]
}

func newModifySchemaHandler[C schemaCommand](repository *repositories.SubjectAttributeSchemaRepository, act schemaAction[C]) cqrs.CommandHandler {
	return &modifySchemaHandler[C]{repository: repository, act: act}
}

func attributeAction[C schemaCommand](
	attributeID func(C) string,
	act func(*aggregates.SubjectAttributeSchema, valueobjects.AttributeID, valueobjects.UserEmail) error,
) schemaAction[C] {
	return func(schema *aggregates.SubjectAttributeSchema, cmd C, modifiedBy valueobjects.UserEmail) (string, error) {
		id, err := valueobjects.NewAttributeIDFromString(attributeID(cmd))
		if err != nil {
			return "", err
		}
		return "", act(schema, id, modifiedBy)
	}
}

func (h *modifySchemaHandler[C]) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(C)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}
	modifiedBy, err := valueobjects.NewUserEmail(command.ModifiedByEmail())
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	schema, err := h.repository.GetByID(ctx, command.SubjectSchemaID())
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	createdID, err := h.act(schema, command, modifiedBy)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := h.repository.Save(ctx, schema); err != nil {
		return cqrs.EmptyResult(), err
	}
	return cqrs.NewResult(createdID), nil
}
