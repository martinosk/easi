package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	"easi/backend/internal/enterprisearchitecture/domain/aggregates"
	"easi/backend/internal/enterprisearchitecture/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

var (
	ErrDomainCapabilityAlreadyLinked  = errors.New("domain capability is already linked to an enterprise capability")
	ErrAncestorLinkedToDifferent      = errors.New("ancestor capability is linked to a different enterprise capability")
	ErrDescendantLinkedToDifferent    = errors.New("descendant capability is linked to a different enterprise capability")
)

type LinkRepository interface {
	Save(ctx context.Context, link *aggregates.EnterpriseCapabilityLink) error
}

type CapabilityRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.EnterpriseCapability, error)
}

type LinkReadModel interface {
	GetByDomainCapabilityID(ctx context.Context, domainCapabilityID string) (*readmodels.EnterpriseCapabilityLinkDTO, error)
	CheckHierarchyConflict(ctx context.Context, domainCapabilityID string, targetEnterpriseCapabilityID string) (*readmodels.HierarchyConflict, error)
}

type LinkCapabilityHandler struct {
	linkRepository       LinkRepository
	capabilityRepository CapabilityRepository
	linkReadModel        LinkReadModel
}

func NewLinkCapabilityHandler(
	linkRepository LinkRepository,
	capabilityRepository CapabilityRepository,
	linkReadModel LinkReadModel,
) *LinkCapabilityHandler {
	return &LinkCapabilityHandler{
		linkRepository:       linkRepository,
		capabilityRepository: capabilityRepository,
		linkReadModel:        linkReadModel,
	}
}

func (h *LinkCapabilityHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.LinkCapability)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	capability, err := h.capabilityRepository.GetByID(ctx, command.EnterpriseCapabilityID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.validateLinkEligibility(ctx, command); err != nil {
		return cqrs.EmptyResult(), err
	}

	link, err := h.createLink(capability, command)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.linkRepository.Save(ctx, link); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(link.ID()), nil
}

func (h *LinkCapabilityHandler) validateLinkEligibility(ctx context.Context, command *commands.LinkCapability) error {
	existingLink, err := h.linkReadModel.GetByDomainCapabilityID(ctx, command.DomainCapabilityID)
	if err != nil {
		return err
	}
	if existingLink != nil {
		return ErrDomainCapabilityAlreadyLinked
	}

	conflict, err := h.linkReadModel.CheckHierarchyConflict(ctx, command.DomainCapabilityID, command.EnterpriseCapabilityID)
	if err != nil {
		return err
	}
	if conflict != nil {
		return h.hierarchyConflictError(conflict)
	}
	return nil
}

func (h *LinkCapabilityHandler) hierarchyConflictError(conflict *readmodels.HierarchyConflict) error {
	if conflict.IsAncestor {
		return ErrAncestorLinkedToDifferent
	}
	return ErrDescendantLinkedToDifferent
}

func (h *LinkCapabilityHandler) createLink(capability *aggregates.EnterpriseCapability, command *commands.LinkCapability) (*aggregates.EnterpriseCapabilityLink, error) {
	domainCapabilityID, err := valueobjects.NewDomainCapabilityIDFromString(command.DomainCapabilityID)
	if err != nil {
		return nil, err
	}

	linkedBy, err := valueobjects.NewLinkedBy(command.LinkedBy)
	if err != nil {
		return nil, err
	}

	return aggregates.NewEnterpriseCapabilityLink(capability, domainCapabilityID, linkedBy)
}
