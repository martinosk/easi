package handlers

import (
	"context"
	"fmt"
	"time"

	"easi/backend/internal/platform/application/commands"
	"easi/backend/internal/platform/domain/aggregates"
	"easi/backend/internal/platform/domain/valueobjects"
	"easi/backend/internal/platform/infrastructure/repositories"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type TenantStore interface {
	Create(ctx context.Context, record repositories.TenantRecord) error
}

type CreateTenantHandler struct {
	repository TenantStore
	eventBus   events.EventBus
}

func NewCreateTenantHandler(repository TenantStore, eventBus events.EventBus) *CreateTenantHandler {
	return &CreateTenantHandler{repository: repository, eventBus: eventBus}
}

func (h *CreateTenantHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.CreateTenant)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	registration, err := validateTenantInput(command)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	tenant, err := aggregates.NewTenant(registration)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Create(ctx, buildTenantRecord(registration)); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.publishTenantCreated(ctx, registration.ID, tenant.GetUncommittedChanges()); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(registration.ID.Value()), nil
}

func validateTenantInput(cmd *commands.CreateTenant) (aggregates.TenantRegistration, error) {
	tenantID, err := sharedvo.NewTenantID(cmd.ID)
	if err != nil {
		return aggregates.TenantRegistration{}, err
	}

	name, err := valueobjects.NewTenantName(cmd.Name)
	if err != nil {
		return aggregates.TenantRegistration{}, err
	}

	domains, err := valueobjects.NewEmailDomainList(cmd.Domains)
	if err != nil {
		return aggregates.TenantRegistration{}, err
	}

	oidcConfig, err := valueobjects.NewOIDCConfig(
		cmd.DiscoveryURL,
		cmd.ClientID,
		valueobjects.OIDCAuthMethod(cmd.AuthMethod),
		cmd.Scopes,
	)
	if err != nil {
		return aggregates.TenantRegistration{}, err
	}

	return aggregates.TenantRegistration{
		ID:              tenantID,
		Name:            name,
		Domains:         domains,
		OIDCConfig:      oidcConfig,
		FirstAdminEmail: cmd.FirstAdminEmail,
	}, nil
}

func buildTenantRecord(registration aggregates.TenantRegistration) repositories.TenantRecord {
	now := time.Now().UTC()
	return repositories.TenantRecord{
		ID:              registration.ID.Value(),
		Name:            registration.Name.Value(),
		Status:          valueobjects.TenantStatusActive.Value(),
		Domains:         registration.DomainNames(),
		DiscoveryURL:    registration.OIDCConfig.DiscoveryURL(),
		ClientID:        registration.OIDCConfig.ClientID(),
		AuthMethod:      string(registration.OIDCConfig.AuthMethod()),
		Scopes:          registration.OIDCConfig.Scopes(),
		FirstAdminEmail: registration.FirstAdminEmail,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func (h *CreateTenantHandler) publishTenantCreated(ctx context.Context, tenantID sharedvo.TenantID, evts []domain.DomainEvent) error {
	if err := h.eventBus.Publish(sharedctx.WithTenant(ctx, tenantID), evts); err != nil {
		return fmt.Errorf("publish TenantCreated %s: %w", tenantID.Value(), err)
	}
	return nil
}
