package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

var (
	ErrParentCapabilityNotFound = errors.New("parent capability not found")
)

type ChangeCapabilityParentHandler struct {
	repository *repositories.CapabilityRepository
	readModel  *readmodels.CapabilityReadModel
}

func NewChangeCapabilityParentHandler(
	repository *repositories.CapabilityRepository,
	readModel *readmodels.CapabilityReadModel,
) *ChangeCapabilityParentHandler {
	return &ChangeCapabilityParentHandler{
		repository: repository,
		readModel:  readModel,
	}
}

func (h *ChangeCapabilityParentHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.ChangeCapabilityParent)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	capability, err := h.repository.GetByID(ctx, command.CapabilityID)
	if err != nil {
		return err
	}

	newParentID, newLevel, err := h.determineNewParentAndLevel(ctx, command)
	if err != nil {
		return err
	}

	if err := h.validateDepthConstraints(ctx, command.CapabilityID, newLevel); err != nil {
		return err
	}

	if err := capability.ChangeParent(newParentID, newLevel); err != nil {
		return err
	}

	if err := h.repository.Save(ctx, capability); err != nil {
		return err
	}

	return h.updateDescendantLevels(ctx, command.CapabilityID, newLevel)
}

func (h *ChangeCapabilityParentHandler) determineNewParentAndLevel(ctx context.Context, command *commands.ChangeCapabilityParent) (valueobjects.CapabilityID, valueobjects.CapabilityLevel, error) {
	if command.NewParentID == "" {
		return valueobjects.CapabilityID{}, valueobjects.LevelL1, nil
	}

	newParentID, err := valueobjects.NewCapabilityIDFromString(command.NewParentID)
	if err != nil {
		return valueobjects.CapabilityID{}, "", err
	}

	parent, err := h.repository.GetByID(ctx, command.NewParentID)
	if err != nil {
		if errors.Is(err, repositories.ErrCapabilityNotFound) {
			return valueobjects.CapabilityID{}, "", ErrParentCapabilityNotFound
		}
		return valueobjects.CapabilityID{}, "", err
	}

	if err := h.detectCircularReference(ctx, command.CapabilityID, command.NewParentID); err != nil {
		return valueobjects.CapabilityID{}, "", err
	}

	newLevel, err := h.calculateChildLevel(parent.Level())
	if err != nil {
		return valueobjects.CapabilityID{}, "", err
	}

	return newParentID, newLevel, nil
}

func (h *ChangeCapabilityParentHandler) validateDepthConstraints(ctx context.Context, capabilityID string, newLevel valueobjects.CapabilityLevel) error {
	subtreeDepth, err := h.calculateSubtreeDepth(ctx, capabilityID)
	if err != nil {
		return err
	}

	if newLevel.NumericValue()+subtreeDepth > 4 {
		return aggregates.ErrWouldExceedMaximumDepth
	}

	return nil
}

func (h *ChangeCapabilityParentHandler) detectCircularReference(ctx context.Context, capabilityID, newParentID string) error {
	currentID := newParentID
	visited := make(map[string]bool)

	for currentID != "" {
		if currentID == capabilityID {
			return aggregates.ErrWouldCreateCircularReference
		}

		if visited[currentID] {
			break
		}
		visited[currentID] = true

		nextParentID, err := h.getParentIDOrBreak(ctx, currentID)
		if err != nil {
			return err
		}

		currentID = nextParentID
	}

	return nil
}

func (h *ChangeCapabilityParentHandler) getParentIDOrBreak(ctx context.Context, capabilityID string) (string, error) {
	parent, err := h.repository.GetByID(ctx, capabilityID)
	if err != nil {
		if errors.Is(err, repositories.ErrCapabilityNotFound) {
			return "", nil
		}
		return "", err
	}
	return parent.ParentID().Value(), nil
}

func (h *ChangeCapabilityParentHandler) calculateChildLevel(parentLevel valueobjects.CapabilityLevel) (valueobjects.CapabilityLevel, error) {
	switch parentLevel {
	case valueobjects.LevelL1:
		return valueobjects.LevelL2, nil
	case valueobjects.LevelL2:
		return valueobjects.LevelL3, nil
	case valueobjects.LevelL3:
		return valueobjects.LevelL4, nil
	default:
		return "", aggregates.ErrWouldExceedMaximumDepth
	}
}

func (h *ChangeCapabilityParentHandler) calculateSubtreeDepth(ctx context.Context, capabilityID string) (int, error) {
	children, err := h.readModel.GetChildren(ctx, capabilityID)
	if err != nil {
		return 0, err
	}

	if len(children) == 0 {
		return 0, nil
	}

	maxChildDepth := 0
	for _, child := range children {
		childDepth, err := h.calculateSubtreeDepth(ctx, child.ID)
		if err != nil {
			return 0, err
		}
		if childDepth > maxChildDepth {
			maxChildDepth = childDepth
		}
	}

	return 1 + maxChildDepth, nil
}

func (h *ChangeCapabilityParentHandler) updateDescendantLevels(ctx context.Context, parentID string, parentLevel valueobjects.CapabilityLevel) error {
	children, err := h.readModel.GetChildren(ctx, parentID)
	if err != nil {
		return err
	}

	childLevel, err := h.calculateChildLevel(parentLevel)
	if err != nil {
		return nil
	}

	for _, child := range children {
		childCapability, err := h.repository.GetByID(ctx, child.ID)
		if err != nil {
			return err
		}

		childParentID, _ := valueobjects.NewCapabilityIDFromString(parentID)
		if err := childCapability.ChangeParent(childParentID, childLevel); err != nil {
			return err
		}

		if err := h.repository.Save(ctx, childCapability); err != nil {
			return err
		}

		if err := h.updateDescendantLevels(ctx, child.ID, childLevel); err != nil {
			return err
		}
	}

	return nil
}
