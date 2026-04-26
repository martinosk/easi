import { useCallback } from 'react';
import { useAppStore } from '../../../store/appStore';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import type { EntityType } from '../utils/dynamicMode';

export function useDraftRemoveFromView() {
  const { currentViewId } = useCurrentView();
  const dynamicViewId = useAppStore((s) => s.dynamicViewId);
  const draftRemoveEntities = useAppStore((s) => s.draftRemoveEntities);
  const draftActive = dynamicViewId !== null && dynamicViewId === currentViewId;

  return useCallback(
    (id: string, _type: EntityType): boolean => {
      if (!draftActive) return false;
      draftRemoveEntities([id]);
      return true;
    },
    [draftActive, draftRemoveEntities],
  );
}
