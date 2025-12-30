package handlers

import (
	"context"
	"time"

	authCommands "easi/backend/internal/auth/application/commands"
	"easi/backend/internal/platform/application/commands"
	"easi/backend/internal/platform/domain/aggregates"
	"easi/backend/internal/platform/domain/valueobjects"
	"easi/backend/internal/platform/infrastructure/repositories"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type CreateTenantHandler struct {
	repository *repositories.TenantRepository
	commandBus cqrs.CommandBus
}

func NewCreateTenantHandler(repository *repositories.TenantRepository, commandBus cqrs.CommandBus) *CreateTenantHandler {
	return &CreateTenantHandler{repository: repository, commandBus: commandBus}
}

func (h *CreateTenantHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.CreateTenant)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	tenantID, err := sharedvo.NewTenantID(command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	name, err := valueobjects.NewTenantName(command.Name)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	domains, err := valueobjects.NewEmailDomainList(command.Domains)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	oidcConfig, err := valueobjects.NewOIDCConfig(
		command.DiscoveryURL,
		command.ClientID,
		valueobjects.OIDCAuthMethod(command.AuthMethod),
		command.Scopes,
	)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	_, err = aggregates.NewTenant(tenantID, name, domains, oidcConfig, command.FirstAdminEmail)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	domainStrings := make([]string, len(domains))
	for i, d := range domains {
		domainStrings[i] = d.Value()
	}

	now := time.Now().UTC()
	record := repositories.TenantRecord{
		ID:              tenantID.Value(),
		Name:            name.Value(),
		Status:          "active",
		Domains:         domainStrings,
		DiscoveryURL:    oidcConfig.DiscoveryURL(),
		ClientID:        oidcConfig.ClientID(),
		AuthMethod:      string(oidcConfig.AuthMethod()),
		Scopes:          oidcConfig.Scopes(),
		FirstAdminEmail: command.FirstAdminEmail,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := h.repository.Create(ctx, record); err != nil {
		return cqrs.EmptyResult(), err
	}

	tenantCtx := sharedctx.WithTenant(ctx, tenantID)
	createInvitationCmd := &authCommands.CreateInvitation{
		Email: command.FirstAdminEmail,
		Role:  "admin",
	}

	if _, err := h.commandBus.Dispatch(tenantCtx, createInvitationCmd); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(tenantID.Value()), nil
}
