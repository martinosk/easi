package handlers

import (
	"context"
	"time"

	authPL "easi/backend/internal/auth/publishedlanguage"
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

	input, err := validateTenantInput(command)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if _, err := aggregates.NewTenant(input.tenantID, input.name, input.domains, input.oidcConfig, command.FirstAdminEmail); err != nil {
		return cqrs.EmptyResult(), err
	}

	record := buildTenantRecord(input, command.FirstAdminEmail)
	if err := h.repository.Create(ctx, record); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.inviteFirstAdmin(ctx, input.tenantID, command.FirstAdminEmail); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(input.tenantID.Value()), nil
}

type validatedTenantInput struct {
	tenantID   sharedvo.TenantID
	name       valueobjects.TenantName
	domains    []valueobjects.EmailDomain
	oidcConfig valueobjects.OIDCConfig
}

func validateTenantInput(cmd *commands.CreateTenant) (validatedTenantInput, error) {
	tenantID, err := sharedvo.NewTenantID(cmd.ID)
	if err != nil {
		return validatedTenantInput{}, err
	}

	name, err := valueobjects.NewTenantName(cmd.Name)
	if err != nil {
		return validatedTenantInput{}, err
	}

	domains, err := valueobjects.NewEmailDomainList(cmd.Domains)
	if err != nil {
		return validatedTenantInput{}, err
	}

	oidcConfig, err := valueobjects.NewOIDCConfig(
		cmd.DiscoveryURL,
		cmd.ClientID,
		valueobjects.OIDCAuthMethod(cmd.AuthMethod),
		cmd.Scopes,
	)
	if err != nil {
		return validatedTenantInput{}, err
	}

	return validatedTenantInput{
		tenantID:   tenantID,
		name:       name,
		domains:    domains,
		oidcConfig: oidcConfig,
	}, nil
}

func buildTenantRecord(input validatedTenantInput, firstAdminEmail string) repositories.TenantRecord {
	domainStrings := make([]string, len(input.domains))
	for i, d := range input.domains {
		domainStrings[i] = d.Value()
	}

	now := time.Now().UTC()
	return repositories.TenantRecord{
		ID:              input.tenantID.Value(),
		Name:            input.name.Value(),
		Status:          "active",
		Domains:         domainStrings,
		DiscoveryURL:    input.oidcConfig.DiscoveryURL(),
		ClientID:        input.oidcConfig.ClientID(),
		AuthMethod:      string(input.oidcConfig.AuthMethod()),
		Scopes:          input.oidcConfig.Scopes(),
		FirstAdminEmail: firstAdminEmail,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func (h *CreateTenantHandler) inviteFirstAdmin(ctx context.Context, tenantID sharedvo.TenantID, email string) error {
	tenantCtx := sharedctx.WithTenant(ctx, tenantID)
	cmd := &authPL.CreateInvitation{
		Email: email,
		Role:  "admin",
	}
	_, err := h.commandBus.Dispatch(tenantCtx, cmd)
	return err
}
