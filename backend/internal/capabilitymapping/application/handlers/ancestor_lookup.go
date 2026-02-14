package handlers

import (
	"context"

	"easi/backend/internal/capabilitymapping/application/readmodels"
)

type CapabilityParentLookup interface {
	GetByID(ctx context.Context, id string) (*readmodels.CapabilityDTO, error)
}

func CollectAncestorIDs(ctx context.Context, lookup CapabilityParentLookup, startID string) ([]string, error) {
	if startID == "" {
		return nil, nil
	}

	ids := []string{}
	visited := map[string]struct{}{}
	currentID := startID

	for currentID != "" {
		if _, seen := visited[currentID]; seen {
			break
		}
		visited[currentID] = struct{}{}
		ids = append(ids, currentID)

		capability, err := lookup.GetByID(ctx, currentID)
		if err != nil {
			return nil, err
		}
		if capability == nil {
			break
		}
		currentID = capability.ParentID
	}

	return ids, nil
}
