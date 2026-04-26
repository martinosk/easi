import { useCallback } from 'react';
import { useAppStore } from '../../../store/appStore';
import { useCapabilities, useRealizations } from '../../capabilities/hooks/useCapabilities';
import { useOriginRelationshipsQuery } from '../../origin-entities/hooks/useOriginRelationships';
import { useRelations } from '../../relations/hooks/useRelations';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { computeOrphans, type EntityType } from '../utils/dynamicMode';

const CASCADE_CONFIRM_THRESHOLD = 5;

export function useDraftRemoveFromView() {
  const { data: relations = [] } = useRelations();
  const { data: capabilities = [] } = useCapabilities();
  const { data: realizations = [] } = useRealizations();
  const { data: originRelationships = [] } = useOriginRelationshipsQuery();
  const { currentViewId } = useCurrentView();

  const dynamicViewId = useAppStore((s) => s.dynamicViewId);
  const dynamicEntities = useAppStore((s) => s.dynamicEntities);
  const dynamicFilters = useAppStore((s) => s.dynamicFilters);
  const draftRemoveEntities = useAppStore((s) => s.draftRemoveEntities);
  const draftActive = dynamicViewId !== null && dynamicViewId === currentViewId;

  return useCallback(
    (id: string, _type: EntityType): boolean => {
      if (!draftActive) return false;
      const data = { relations, capabilities, realizations, originRelationships };
      const orphans = computeOrphans(data, dynamicEntities, id, dynamicFilters);
      if (orphans.length >= CASCADE_CONFIRM_THRESHOLD) {
        const total = 1 + orphans.length;
        const plural = orphans.length === 1 ? '' : 's';
        const proceed = window.confirm(
          `Removing this entity will also remove ${orphans.length} orphaned descendant${plural} (${total} total). Continue?`,
        );
        if (!proceed) return false;
      }
      draftRemoveEntities([id, ...orphans]);
      return true;
    },
    [
      draftActive,
      relations,
      capabilities,
      realizations,
      originRelationships,
      dynamicEntities,
      dynamicFilters,
      draftRemoveEntities,
    ],
  );
}
