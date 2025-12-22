import { useCallback } from 'react';
import { useAppStore } from '../store/appStore';
import { useRemoveComponentFromView, useAddComponentToView } from '../features/views/hooks/useViews';
import { useCurrentView } from './useCurrentView';
import type { ComponentId, ViewId } from '../api/types';

export function useViewOperations() {
  const { currentViewId } = useCurrentView();
  const setCurrentViewId = useAppStore((state) => state.setCurrentViewId);
  const clearSelection = useAppStore((state) => state.clearSelection);

  const removeComponentMutation = useRemoveComponentFromView();
  const addComponentMutation = useAddComponentToView();

  const removeComponentFromView = useCallback(
    async (componentId: ComponentId) => {
      if (!currentViewId) {
        console.warn('No current view selected');
        return;
      }

      try {
        await removeComponentMutation.mutateAsync({
          viewId: currentViewId,
          componentId,
        });
        clearSelection();
      } catch (error) {
        console.error('Failed to remove component from view:', error);
      }
    },
    [currentViewId, removeComponentMutation, clearSelection]
  );

  const addComponentToView = useCallback(
    async (componentId: ComponentId, x: number, y: number) => {
      if (!currentViewId) {
        console.warn('No current view selected');
        return;
      }

      try {
        await addComponentMutation.mutateAsync({
          viewId: currentViewId,
          request: { componentId, x, y },
        });
      } catch (error) {
        console.error('Failed to add component to view:', error);
      }
    },
    [currentViewId, addComponentMutation]
  );

  const switchView = useCallback((viewId: ViewId) => {
    setCurrentViewId(viewId);
  }, [setCurrentViewId]);

  return {
    removeComponentFromView,
    addComponentToView,
    switchView,
  };
}
