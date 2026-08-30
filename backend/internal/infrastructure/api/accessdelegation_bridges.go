package api

import (
	"context"
	"fmt"
	"log"

	acdPorts "easi/backend/internal/accessdelegation/application/ports"
	accessdelegationAPI "easi/backend/internal/accessdelegation/infrastructure/api"
	adServices "easi/backend/internal/accessdelegation/infrastructure/services"
	archReadModels "easi/backend/internal/architecturemodeling/application/readmodels"
	archAdapters "easi/backend/internal/architecturemodeling/infrastructure/adapters"
	viewReadModels "easi/backend/internal/architectureviews/application/readmodels"
	viewAdapters "easi/backend/internal/architectureviews/infrastructure/adapters"
	authCommands "easi/backend/internal/auth/application/commands"
	authReadModels "easi/backend/internal/auth/application/readmodels"
	authAdapters "easi/backend/internal/auth/infrastructure/adapters"
	capReadModels "easi/backend/internal/capabilitymapping/application/readmodels"
	capAdapters "easi/backend/internal/capabilitymapping/infrastructure/adapters"
	"easi/backend/internal/shared/cqrs"
)

func accessDelegationRoutesDeps(deps routerDependencies) accessdelegationAPI.AccessDelegationRoutesDeps {
	return accessdelegationAPI.AccessDelegationRoutesDeps{
		CommandBus:     deps.commandBus,
		EventStore:     deps.eventStore,
		EventBus:       deps.eventBus,
		DB:             deps.db,
		HATEOAS:        deps.hateoas,
		AuthMiddleware: deps.authDeps.AuthMiddleware,
		NameLookups: adServices.ArtifactNameResolverDeps{
			Capabilities:     capAdapters.NewCapabilityNameAdapter(capReadModels.NewCapabilityReadModel(deps.db)),
			Components:       archAdapters.NewComponentNameAdapter(archReadModels.NewApplicationComponentReadModel(deps.db)),
			Views:            viewAdapters.NewViewNameAdapter(viewReadModels.NewArchitectureViewReadModel(deps.db)),
			Domains:          capAdapters.NewDomainNameAdapter(capReadModels.NewBusinessDomainReadModel(deps.db)),
			Vendors:          archAdapters.NewVendorNameAdapter(archReadModels.NewVendorReadModel(deps.db)),
			AcquiredEntities: archAdapters.NewAcquiredEntityNameAdapter(archReadModels.NewAcquiredEntityReadModel(deps.db)),
			InternalTeams:    archAdapters.NewInternalTeamNameAdapter(archReadModels.NewInternalTeamReadModel(deps.db)),
		},
		UserLookup:    authAdapters.NewUserEmailLookupAdapter(authReadModels.NewUserReadModel(deps.db)),
		InvChecker:    authAdapters.NewInvitationCheckerAdapter(authReadModels.NewInvitationReadModel(deps.db)),
		DomainChecker: authAdapters.NewDomainAllowlistCheckerAdapter(authReadModels.NewTenantDomainChecker(deps.db)),
		Invitations:   invitationRequester{commandBus: deps.commandBus},
	}
}

const editGrantInviteeRole = "stakeholder"

type invitationRequester struct {
	commandBus cqrs.CommandBus
}

func (a invitationRequester) RequestInvitation(ctx context.Context, request acdPorts.InvitationRequest) error {
	cmd := &authCommands.CreateInvitation{
		Email:        request.GranteeEmail,
		Role:         editGrantInviteeRole,
		InviterID:    request.GrantorID,
		InviterEmail: request.GrantorEmail,
	}
	if _, err := a.commandBus.Dispatch(ctx, cmd); err != nil {
		return fmt.Errorf("create invitation for %s: %w", request.GranteeEmail, err)
	}
	log.Printf("[AUDIT] invitation-auto-created email=%s inviter=%s reason=edit-grant", request.GranteeEmail, request.GrantorEmail)
	return nil
}
