import { useMemo } from 'react';
import { useAppStore } from '../../../store/appStore';
import { type EntityIdSets, entityIdSets, viewToEntityRefs } from '../utils/viewElements';
import { useCurrentView } from './useCurrentView';

export function useCurrentViewElementIds(): EntityIdSets {
  const { currentView, currentViewId } = useCurrentView();
  const dynamicViewId = useAppStore((s) => s.dynamicViewId);
  const dynamicEntities = useAppStore((s) => s.dynamicEntities);

  return useMemo(() => {
    if (currentViewId && dynamicViewId === currentViewId) {
      return entityIdSets(dynamicEntities);
    }
    return entityIdSets(viewToEntityRefs(currentView));
  }, [currentView, currentViewId, dynamicViewId, dynamicEntities]);
}
