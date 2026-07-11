package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/onepagers/application/commands"
	"easi/backend/internal/onepagers/domain/aggregates"
	"easi/backend/internal/onepagers/domain/valueobjects"
	"easi/backend/internal/onepagers/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

var ErrConfigurationAlreadyExists = errors.New("a one-pager configuration already exists for this subject type")

type ConfigurationLookup interface {
	ConfigurationExists(ctx context.Context, subjectType string) (bool, error)
}

type CreateOnePagerConfigurationHandler struct {
	repository *repositories.OnePagerConfigurationRepository
	lookup     ConfigurationLookup
}

func NewCreateOnePagerConfigurationHandler(
	repository *repositories.OnePagerConfigurationRepository,
	lookup ConfigurationLookup,
) *CreateOnePagerConfigurationHandler {
	return &CreateOnePagerConfigurationHandler{repository: repository, lookup: lookup}
}

func (h *CreateOnePagerConfigurationHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.CreateOnePagerConfiguration)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	input, err := parseCreateConfiguration(command)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	exists, err := h.lookup.ConfigurationExists(ctx, input.subjectType.Value())
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if exists {
		return cqrs.EmptyResult(), ErrConfigurationAlreadyExists
	}

	config, err := aggregates.NewOnePagerConfiguration(input.tenantID, input.subjectType, input.createdBy)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, config); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(config.ID()), nil
}

type createConfigurationInput struct {
	tenantID    sharedvo.TenantID
	subjectType valueobjects.SubjectType
	createdBy   valueobjects.UserEmail
}

func parseCreateConfiguration(command *commands.CreateOnePagerConfiguration) (createConfigurationInput, error) {
	tenantID, err := sharedvo.NewTenantID(command.TenantID)
	if err != nil {
		return createConfigurationInput{}, err
	}
	subjectType, err := valueobjects.NewSubjectType(command.SubjectType)
	if err != nil {
		return createConfigurationInput{}, err
	}
	createdBy, err := valueobjects.NewUserEmail(command.CreatedBy)
	if err != nil {
		return createConfigurationInput{}, err
	}
	return createConfigurationInput{tenantID: tenantID, subjectType: subjectType, createdBy: createdBy}, nil
}
