package handlers

import (
	"context"
	"fmt"
	"time"

	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/platform/application/commands"
	"easi/backend/internal/platform/domain/aggregates"
	"easi/backend/internal/platform/domain/valueobjects"
	"easi/backend/internal/platform/infrastructure/repositories"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type TenantStore interface {
	Create(ctx context.Context, record repositories.TenantRecord) error
}

type CreateTenantHandler struct {
	repository TenantStore
	eventStore eventstore.EventStore
}

func NewCreateTenantHandler(repository TenantStore, eventStore eventstore.EventStore) *CreateTenantHandler {
	return &CreateTenantHandler{repository: repository, eventStore: eventStore}
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

	tenantCtx := sharedctx.WithTenant(ctx, registration.ID)

	if err := h.repository.Create(tenantCtx, buildTenantRecord(registration)); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.saveTenantCreated(tenantCtx, registration.ID, tenant); err != nil {
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

func (h *CreateTenantHandler) saveTenantCreated(ctx context.Context, tenantID sharedvo.TenantID, tenant *aggregates.Tenant) error {
	changes := tenant.GetUncommittedChanges()
	expectedVersion := tenant.Version() - len(changes)

	if err := h.eventStore.SaveEvents(ctx, tenantID.Value(), changes, expectedVersion); err != nil {
		return fmt.Errorf("save TenantCreated %s: %w", tenantID.Value(), err)
	}
	tenant.MarkChangesAsCommitted()
	return nil
}
