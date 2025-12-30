package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	"easi/backend/internal/enterprisearchitecture/domain/aggregates"
	"easi/backend/internal/enterprisearchitecture/domain/valueobjects"
	"easi/backend/internal/enterprisearchitecture/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

var ErrDomainCapabilityAlreadyLinked = errors.New("domain capability is already linked to an enterprise capability")

type LinkCapabilityHandler struct {
	linkRepository       *repositories.EnterpriseCapabilityLinkRepository
	capabilityRepository *repositories.EnterpriseCapabilityRepository
	linkReadModel        *readmodels.EnterpriseCapabilityLinkReadModel
}

func NewLinkCapabilityHandler(
	linkRepository *repositories.EnterpriseCapabilityLinkRepository,
	capabilityRepository *repositories.EnterpriseCapabilityRepository,
	linkReadModel *readmodels.EnterpriseCapabilityLinkReadModel,
) *LinkCapabilityHandler {
	return &LinkCapabilityHandler{
		linkRepository:       linkRepository,
		capabilityRepository: capabilityRepository,
		linkReadModel:        linkReadModel,
	}
}

func (h *LinkCapabilityHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.LinkCapability)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	capability, err := h.capabilityRepository.GetByID(ctx, command.EnterpriseCapabilityID)
	if err != nil {
		return err
	}

	existingLink, err := h.linkReadModel.GetByDomainCapabilityID(ctx, command.DomainCapabilityID)
	if err != nil {
		return err
	}
	if existingLink != nil {
		return ErrDomainCapabilityAlreadyLinked
	}

	domainCapabilityID, err := valueobjects.NewDomainCapabilityIDFromString(command.DomainCapabilityID)
	if err != nil {
		return err
	}

	linkedBy, err := valueobjects.NewLinkedBy(command.LinkedBy)
	if err != nil {
		return err
	}

	link, err := aggregates.NewEnterpriseCapabilityLink(capability, domainCapabilityID, linkedBy)
	if err != nil {
		return err
	}

	command.ID = link.ID()

	return h.linkRepository.Save(ctx, link)
}
