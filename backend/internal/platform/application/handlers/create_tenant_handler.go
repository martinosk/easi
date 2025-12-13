package handlers

import (
	"context"
	"time"

	"easi/backend/internal/platform/application/commands"
	"easi/backend/internal/platform/domain/aggregates"
	"easi/backend/internal/platform/domain/valueobjects"
	"easi/backend/internal/platform/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
	sharedvo "easi/backend/internal/shared/domain/valueobjects"
)

type CreateTenantHandler struct {
	repository *repositories.TenantRepository
}

func NewCreateTenantHandler(repository *repositories.TenantRepository) *CreateTenantHandler {
	return &CreateTenantHandler{repository: repository}
}

func (h *CreateTenantHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.CreateTenant)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	tenantID, err := sharedvo.NewTenantID(command.ID)
	if err != nil {
		return err
	}

	name, err := valueobjects.NewTenantName(command.Name)
	if err != nil {
		return err
	}

	domains, err := valueobjects.NewEmailDomainList(command.Domains)
	if err != nil {
		return err
	}

	oidcConfig, err := valueobjects.NewOIDCConfig(
		command.DiscoveryURL,
		command.ClientID,
		valueobjects.OIDCAuthMethod(command.AuthMethod),
		command.Scopes,
	)
	if err != nil {
		return err
	}

	_, err = aggregates.NewTenant(tenantID, name, domains, oidcConfig, command.FirstAdminEmail)
	if err != nil {
		return err
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

	return h.repository.Create(ctx, record)
}
