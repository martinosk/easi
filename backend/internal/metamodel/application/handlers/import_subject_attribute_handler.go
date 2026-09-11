package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/metamodel/application/commands"
	"easi/backend/internal/metamodel/domain/aggregates"
	"easi/backend/internal/metamodel/domain/valueobjects"
	"easi/backend/internal/metamodel/infrastructure/repositories"
	mmPL "easi/backend/internal/metamodel/publishedlanguage"
	"easi/backend/internal/shared/cqrs"
)

type ImportSubjectAttributeHandler struct {
	repository *repositories.SubjectAttributeSchemaRepository
	lookup     SchemaLookup
}

func NewImportSubjectAttributeHandler(repository *repositories.SubjectAttributeSchemaRepository, lookup SchemaLookup) *ImportSubjectAttributeHandler {
	return &ImportSubjectAttributeHandler{repository: repository, lookup: lookup}
}

func (h *ImportSubjectAttributeHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.ImportSubjectAttribute)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}
	importedBy, err := valueobjects.NewUserEmail(command.ImportedBy)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	attribute, err := importedAttribute(command.Attribute)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	schema, err := h.ensureSchema(ctx, command)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := schema.ImportAttribute(attribute, importedBy); err != nil {
		if errors.Is(err, aggregates.ErrAttributeAlreadyDefined) {
			return cqrs.NewResult(schema.ID()), nil
		}
		return cqrs.EmptyResult(), err
	}
	if err := h.repository.Save(ctx, schema); err != nil {
		return cqrs.EmptyResult(), err
	}
	return cqrs.NewResult(schema.ID()), nil
}

func (h *ImportSubjectAttributeHandler) ensureSchema(ctx context.Context, command *commands.ImportSubjectAttribute) (*aggregates.SubjectAttributeSchema, error) {
	existing, err := h.lookup.GetBySubjectType(ctx, command.SubjectType)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return h.repository.GetByID(ctx, existing.ID)
	}
	schema, err := newSchema(command.TenantID, command.SubjectType, command.ImportedBy)
	if err != nil {
		return nil, err
	}
	if err := h.repository.Save(ctx, schema); err != nil {
		return nil, err
	}
	return schema, nil
}

func importedAttribute(imported mmPL.ImportedAttribute) (valueobjects.SubjectAttribute, error) {
	id, err := valueobjects.NewAttributeIDFromString(imported.ID)
	if err != nil {
		return valueobjects.SubjectAttribute{}, err
	}
	name, err := valueobjects.NewAttributeName(imported.Name)
	if err != nil {
		return valueobjects.SubjectAttribute{}, err
	}
	attributeType, err := valueobjects.NewAttributeType(imported.Type)
	if err != nil {
		return valueobjects.SubjectAttribute{}, err
	}
	helpText, err := valueobjects.NewHelpText(imported.HelpText)
	if err != nil {
		return valueobjects.SubjectAttribute{}, err
	}
	options, err := importedOptions(imported.Options)
	if err != nil {
		return valueobjects.SubjectAttribute{}, err
	}
	attribute, err := valueobjects.NewSubjectAttribute(valueobjects.SubjectAttributeParams{
		ID: id, Name: name, Type: attributeType, HelpText: helpText, Options: options, Min: imported.Min, Max: imported.Max,
	})
	if err != nil {
		return valueobjects.SubjectAttribute{}, err
	}
	if !imported.Active {
		attribute = attribute.Retired()
	}
	return attribute, nil
}

func importedOptions(imported []mmPL.ImportedOption) ([]valueobjects.AttributeOption, error) {
	options := make([]valueobjects.AttributeOption, len(imported))
	for i, option := range imported {
		optionID, err := valueobjects.NewOptionIDFromString(option.ID)
		if err != nil {
			return nil, err
		}
		label, err := valueobjects.NewOptionLabel(option.Label)
		if err != nil {
			return nil, err
		}
		if option.Active {
			options[i] = valueobjects.NewAttributeOption(optionID, label)
		} else {
			options[i] = valueobjects.NewRetiredAttributeOption(optionID, label)
		}
	}
	return options, nil
}
