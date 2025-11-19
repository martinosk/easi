import type { StateCreator } from 'zustand';
import type { Component, Relation, View } from '../../api/types';
import type { ComponentId, Position, EdgeType, LayoutDirection } from '../types/storeTypes';
import apiClient from '../../api/client';
import { handleApiCall, optimisticUpdate } from '../utils/apiHelpers';
import toast from 'react-hot-toast';
import { ApiError } from '../../api/types';

export interface LayoutActions {
  updatePosition: (componentId: ComponentId, position: Position) => Promise<void>;
  setEdgeType: (edgeType: EdgeType) => Promise<void>;
  setLayoutDirection: (direction: LayoutDirection) => Promise<void>;
  applyAutoLayout: () => Promise<void>;
}

type StoreWithDependencies = {
  currentView: View | null;
  components: Component[];
  relations: Relation[];
};

export const createLayoutSlice: StateCreator<
  StoreWithDependencies & LayoutActions,
  [],
  [],
  LayoutActions
> = (set, get) => ({
  updatePosition: async (componentId: ComponentId, position: Position) => {
    const { currentView } = get();

    if (!currentView) {
      return;
    }

    await handleApiCall(
      () => apiClient.updateComponentPosition(currentView.id, componentId, position),
      'Failed to update position'
    );

    const updatedComponents = currentView.components.map((vc) =>
      vc.componentId === componentId ? { ...vc, ...position } : vc
    );

    set({
      currentView: {
        ...currentView,
        components: updatedComponents,
      },
    });
  },

  setEdgeType: async (edgeType: EdgeType) => {
    const { currentView } = get();

    if (!currentView) {
      return;
    }

    const previousEdgeType = currentView.edgeType;

    set({
      currentView: {
        ...currentView,
        edgeType,
      },
    });

    await optimisticUpdate(
      () => apiClient.updateViewEdgeType(currentView.id, { edgeType }),
      () => {},
      () => set({
        currentView: {
          ...currentView,
          edgeType: previousEdgeType,
        },
      }),
      'Edge type updated',
      'Failed to update edge type'
    );
  },

  setLayoutDirection: async (layoutDirection: LayoutDirection) => {
    const { currentView } = get();

    if (!currentView) {
      return;
    }

    const previousLayoutDirection = currentView.layoutDirection;

    set({
      currentView: {
        ...currentView,
        layoutDirection,
      },
    });

    await optimisticUpdate(
      () => apiClient.updateViewLayoutDirection(currentView.id, { layoutDirection }),
      () => {},
      () => set({
        currentView: {
          ...currentView,
          layoutDirection: previousLayoutDirection,
        },
      }),
      'Layout direction updated',
      'Failed to update layout direction'
    );
  },

  applyAutoLayout: async () => {
    const { currentView, components, relations } = get();

    if (!currentView) {
      return;
    }

    try {
      const { calculateDagreLayout } = await import('../../utils/layout');

      const nodes = components
        .filter((component) =>
          currentView.components.some((vc) => vc.componentId === component.id)
        )
        .map((component) => {
          const viewComponent = currentView.components.find(
            (vc) => vc.componentId === component.id
          );

          const position = viewComponent
            ? { x: viewComponent.x, y: viewComponent.y }
            : { x: 400, y: 300 };

          return {
            id: component.id,
            type: 'component',
            position,
            data: {
              label: component.name,
              description: component.description,
            },
          };
        });

      const edges = relations
        .filter((relation) => {
          const sourceInView = currentView.components.some(
            (vc) => vc.componentId === relation.sourceComponentId
          );
          const targetInView = currentView.components.some(
            (vc) => vc.componentId === relation.targetComponentId
          );
          return sourceInView && targetInView;
        })
        .map((relation) => ({
          id: relation.id,
          source: relation.sourceComponentId,
          target: relation.targetComponentId,
        }));

      const layoutedNodes = calculateDagreLayout(nodes, edges, {
        direction: (currentView.layoutDirection as 'TB' | 'LR' | 'BT' | 'RL') || 'TB',
      });

      const positions = layoutedNodes.map((node) => ({
        componentId: node.id,
        x: node.position.x,
        y: node.position.y,
      }));

      await apiClient.updateMultiplePositions(currentView.id, { positions });

      const updatedView = await apiClient.getViewById(currentView.id);
      set({ currentView: updatedView });

      toast.success('Layout applied');
    } catch (error) {
      const errorMessage = error instanceof ApiError
        ? error.message
        : 'Failed to apply layout';

      toast.error(errorMessage);
      throw error;
    }
  },
});
