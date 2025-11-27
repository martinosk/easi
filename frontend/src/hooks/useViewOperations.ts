import { useCallback } from 'react';
import { useAppStore } from '../store/appStore';
import apiClient from '../api/client';
import type { ComponentId, ViewId } from '../api/types';
import toast from 'react-hot-toast';

export function useViewOperations() {
  const currentView = useAppStore((state) => state.currentView);

  const removeComponentFromView = useCallback(
    async (componentId: ComponentId) => {
      if (!currentView) {
        console.warn('No current view selected');
        return;
      }

      try {
        await apiClient.removeComponentFromView(currentView.id, componentId);
        const updatedView = await apiClient.getViewById(currentView.id);
        useAppStore.setState({ currentView: updatedView });
        useAppStore.getState().clearSelection();
        toast.success('Component removed from view');
      } catch (error) {
        console.error('Failed to remove component from view:', error);
        toast.error('Failed to remove component from view');
      }
    },
    [currentView]
  );

  const addComponentToView = useCallback(
    async (componentId: ComponentId, x: number, y: number) => {
      if (!currentView) {
        console.warn('No current view selected');
        return;
      }

      try {
        await apiClient.addComponentToView(currentView.id, {
          componentId,
          x,
          y,
        });
        const updatedView = await apiClient.getViewById(currentView.id);
        useAppStore.setState({ currentView: updatedView });
        toast.success('Component added to view');
      } catch (error) {
        console.error('Failed to add component to view:', error);
        toast.error('Failed to add component to view');
      }
    },
    [currentView]
  );

  const switchView = useCallback(async (viewId: ViewId) => {
    const switchViewAction = useAppStore.getState().switchView;
    try {
      await switchViewAction(viewId);
    } catch (error) {
      console.error('Failed to switch view:', error);
    }
  }, []);

  return {
    removeComponentFromView,
    addComponentToView,
    switchView,
  };
}
