package handlers

import (
	"easi/backend/internal/metamodel/application/commands"
	"easi/backend/internal/metamodel/domain/aggregates"
	"easi/backend/internal/metamodel/domain/valueobjects"
	"easi/backend/internal/metamodel/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

func NewDefineSubjectAttributeHandler(repository *repositories.SubjectAttributeSchemaRepository) cqrs.CommandHandler {
	return newModifySchemaHandler(repository, defineSubjectAttribute)
}

func NewRenameSubjectAttributeHandler(repository *repositories.SubjectAttributeSchemaRepository) cqrs.CommandHandler {
	return newModifySchemaHandler(repository, renameSubjectAttribute)
}

func NewRetireSubjectAttributeHandler(repository *repositories.SubjectAttributeSchemaRepository) cqrs.CommandHandler {
	return newModifySchemaHandler(repository, attributeAction(
		func(c *commands.RetireSubjectAttribute) string { return c.AttributeID },
		(*aggregates.SubjectAttributeSchema).RetireAttribute,
	))
}

func NewReactivateSubjectAttributeHandler(repository *repositories.SubjectAttributeSchemaRepository) cqrs.CommandHandler {
	return newModifySchemaHandler(repository, attributeAction(
		func(c *commands.ReactivateSubjectAttribute) string { return c.AttributeID },
		(*aggregates.SubjectAttributeSchema).ReactivateAttribute,
	))
}

func NewAddSubjectAttributeOptionHandler(repository *repositories.SubjectAttributeSchemaRepository) cqrs.CommandHandler {
	return newModifySchemaHandler(repository, addSubjectAttributeOption)
}

func NewRetireSubjectAttributeOptionHandler(repository *repositories.SubjectAttributeSchemaRepository) cqrs.CommandHandler {
	return newModifySchemaHandler(repository, retireSubjectAttributeOption)
}

func NewSetSubjectAttributeBoundsHandler(repository *repositories.SubjectAttributeSchemaRepository) cqrs.CommandHandler {
	return newModifySchemaHandler(repository, setSubjectAttributeBounds)
}

func defineSubjectAttribute(schema *aggregates.SubjectAttributeSchema, c *commands.DefineSubjectAttribute, modifiedBy valueobjects.UserEmail) (string, error) {
	params, err := buildDefineParams(c)
	if err != nil {
		return "", err
	}
	attributeID, err := schema.DefineAttribute(params, modifiedBy)
	if err != nil {
		return "", err
	}
	return attributeID.Value(), nil
}

func buildDefineParams(c *commands.DefineSubjectAttribute) (aggregates.DefineAttributeParams, error) {
	name, err := valueobjects.NewAttributeName(c.Name)
	if err != nil {
		return aggregates.DefineAttributeParams{}, err
	}
	attributeType, err := valueobjects.NewAttributeType(c.AttributeType)
	if err != nil {
		return aggregates.DefineAttributeParams{}, err
	}
	helpText, err := valueobjects.NewHelpText(c.HelpText)
	if err != nil {
		return aggregates.DefineAttributeParams{}, err
	}
	labels := make([]valueobjects.OptionLabel, len(c.OptionLabels))
	for i, raw := range c.OptionLabels {
		label, err := valueobjects.NewOptionLabel(raw)
		if err != nil {
			return aggregates.DefineAttributeParams{}, err
		}
		labels[i] = label
	}
	return aggregates.DefineAttributeParams{
		Name: name, Type: attributeType, HelpText: helpText, OptionLabels: labels, Min: c.Min, Max: c.Max,
	}, nil
}

func renameSubjectAttribute(schema *aggregates.SubjectAttributeSchema, c *commands.RenameSubjectAttribute, modifiedBy valueobjects.UserEmail) (string, error) {
	attributeID, err := valueobjects.NewAttributeIDFromString(c.AttributeID)
	if err != nil {
		return "", err
	}
	name, err := valueobjects.NewAttributeName(c.Name)
	if err != nil {
		return "", err
	}
	helpText, err := valueobjects.NewHelpText(c.HelpText)
	if err != nil {
		return "", err
	}
	return "", schema.RenameAttribute(aggregates.RenameAttributeParams{
		AttributeID: attributeID, Name: name, HelpText: helpText, RequestedType: c.RequestedType,
	}, modifiedBy)
}

func addSubjectAttributeOption(schema *aggregates.SubjectAttributeSchema, c *commands.AddSubjectAttributeOption, modifiedBy valueobjects.UserEmail) (string, error) {
	attributeID, err := valueobjects.NewAttributeIDFromString(c.AttributeID)
	if err != nil {
		return "", err
	}
	label, err := valueobjects.NewOptionLabel(c.Label)
	if err != nil {
		return "", err
	}
	optionID, err := schema.AddOption(attributeID, label, modifiedBy)
	if err != nil {
		return "", err
	}
	return optionID.Value(), nil
}

func retireSubjectAttributeOption(schema *aggregates.SubjectAttributeSchema, c *commands.RetireSubjectAttributeOption, modifiedBy valueobjects.UserEmail) (string, error) {
	attributeID, err := valueobjects.NewAttributeIDFromString(c.AttributeID)
	if err != nil {
		return "", err
	}
	optionID, err := valueobjects.NewOptionIDFromString(c.OptionID)
	if err != nil {
		return "", err
	}
	return "", schema.RetireOption(attributeID, optionID, modifiedBy)
}

func setSubjectAttributeBounds(schema *aggregates.SubjectAttributeSchema, c *commands.SetSubjectAttributeBounds, modifiedBy valueobjects.UserEmail) (string, error) {
	attributeID, err := valueobjects.NewAttributeIDFromString(c.AttributeID)
	if err != nil {
		return "", err
	}
	return "", schema.SetBounds(attributeID, c.Min, c.Max, modifiedBy)
}
