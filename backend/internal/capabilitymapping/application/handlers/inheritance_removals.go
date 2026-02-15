package handlers

import (
	"sort"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/events"
)

func BuildRealizationRemovals(realizations []readmodels.RealizationDTO, ancestorIDs []string) []events.RealizationInheritanceRemoval {
	sourceIDs := collectUniqueSourceIDs(realizations)
	if len(sourceIDs) == 0 {
		return nil
	}

	removals := make([]events.RealizationInheritanceRemoval, 0, len(sourceIDs))
	for _, sourceID := range sourceIDs {
		removals = append(removals, events.RealizationInheritanceRemoval{
			SourceRealizationID: sourceID,
			CapabilityIDs:       ancestorIDs,
		})
	}

	return removals
}

func collectUniqueSourceIDs(realizations []readmodels.RealizationDTO) []string {
	seen := make(map[string]struct{})
	for _, realization := range realizations {
		sourceID := resolveSourceRealizationID(realization)
		if sourceID == "" {
			continue
		}
		seen[sourceID] = struct{}{}
	}

	keys := make([]string, 0, len(seen))
	for sourceID := range seen {
		keys = append(keys, sourceID)
	}
	sort.Strings(keys)

	return keys
}
