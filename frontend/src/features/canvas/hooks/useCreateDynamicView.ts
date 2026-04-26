import { useCallback, useState } from 'react';
import toast from 'react-hot-toast';
import { useAppStore } from '../../../store/appStore';
import type { EntityRef } from '../utils/dynamicMode';
import { useCreateView } from '../../views/hooks/useViews';

const MAX_VIEW_NAME_LENGTH = 100;

function buildViewName(entityName: string): string {
  const prefix = 'Dynamic view for ';
  const maxEntityLength = MAX_VIEW_NAME_LENGTH - prefix.length;
  const truncatedName =
    entityName.length > maxEntityLength ? `${entityName.slice(0, maxEntityLength - 1)}…` : entityName;
  return `${prefix}${truncatedName}`;
}

export function useCreateDynamicView() {
  const createViewMutation = useCreateView();
  const setCurrentViewId = useAppStore((state) => state.setCurrentViewId);
  const openView = useAppStore((state) => state.openView);
  const enterDynamicMode = useAppStore((state) => state.enterDynamicMode);
  const [isCreating, setIsCreating] = useState(false);

  const create = useCallback(
    async (seed: EntityRef, entityName: string) => {
      setIsCreating(true);
      try {
        const newView = await createViewMutation.mutateAsync({ name: buildViewName(entityName) });
        openView(newView.id);
        setCurrentViewId(newView.id);
        enterDynamicMode(
          {
            entities: [seed],
            positions: { [seed.id]: { x: 0, y: 0 } },
          },
          newView.id,
        );
        toast.success(`Dynamic view ready — expand from ${entityName}`);
      } catch {
        toast.error('Failed to create dynamic view');
      } finally {
        setIsCreating(false);
      }
    },
    [createViewMutation, setCurrentViewId, openView, enterDynamicMode],
  );

  return { create, isCreating };
}
