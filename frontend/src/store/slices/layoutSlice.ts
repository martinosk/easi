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
  setColorScheme: (colorScheme: string) => Promise<void>;
  applyAutoLayout: () => Promise<void>;
}

type StoreWithDependencies = {
  currentView: View | null;
  components: Component[];
  relations: Relation[];
  capabilities: import('../../api/types').Capability[];
  capabilityDependencies: import('../../api/types').CapabilityDependency[];
  capabilityRealizations: import('../../api/types').CapabilityRealization[];
  canvasCapabilities: Array<{ capabilityId: string; x: number; y: number }>;
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

  setColorScheme: async (colorScheme: string) => {
    const { currentView } = get();

    if (!currentView) {
      return;
    }

    const previousColorScheme = currentView.colorScheme;

    set({
      currentView: {
        ...currentView,
        colorScheme,
      },
    });

    await optimisticUpdate(
      () => apiClient.updateViewColorScheme(currentView.id, { colorScheme }),
      () => {},
      () => set({
        currentView: {
          ...currentView,
          colorScheme: previousColorScheme,
        },
      }),
      'Color scheme updated',
      'Failed to update color scheme'
    );
  },

  applyAutoLayout: async () => {
    const { currentView, components, relations, capabilities, capabilityDependencies, capabilityRealizations, canvasCapabilities } = get();

    if (!currentView) {
      return;
    }

    try {
      const { calculateDagreLayout } = await import('../../utils/layout');

      const componentNodes = components
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

      const capabilityNodes = canvasCapabilities
        .map((canvasCapability) => {
          const capability = capabilities.find((c) => c.id === canvasCapability.capabilityId);
          if (!capability) return null;

          return {
            id: capability.id,
            type: 'capability',
            position: { x: canvasCapability.x, y: canvasCapability.y },
            data: {
              label: capability.name,
              description: capability.description,
            },
          };
        })
        .filter((node): node is NonNullable<typeof node> => node !== null);

      const nodes = [...componentNodes, ...capabilityNodes];

      const relationEdges = relations
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

      const capabilityParentEdges = canvasCapabilities
        .map((canvasCapability) => {
          const capability = capabilities.find((c) => c.id === canvasCapability.capabilityId);
          if (!capability || !capability.parentId) return null;

          const parentInView = canvasCapabilities.some((cc) => cc.capabilityId === capability.parentId);
          if (!parentInView) return null;

          return {
            id: `parent-${capability.id}`,
            source: capability.parentId,
            target: capability.id,
          };
        })
        .filter((edge): edge is NonNullable<typeof edge> => edge !== null);

      const capabilityDependencyEdges = capabilityDependencies
        .filter((dep) => {
          const sourceInView = canvasCapabilities.some((cc) => cc.capabilityId === dep.sourceCapabilityId);
          const targetInView = canvasCapabilities.some((cc) => cc.capabilityId === dep.targetCapabilityId);
          return sourceInView && targetInView;
        })
        .map((dep) => ({
          id: dep.id,
          source: dep.sourceCapabilityId,
          target: dep.targetCapabilityId,
        }));

      const realizationEdges = capabilityRealizations
        .filter((real) => {
          const capabilityInView = canvasCapabilities.some((cc) => cc.capabilityId === real.capabilityId);
          const componentInView = currentView.components.some((vc) => vc.componentId === real.componentId);
          return capabilityInView && componentInView;
        })
        .map((real) => ({
          id: real.id,
          source: real.componentId,
          target: real.capabilityId,
        }));

      const edges = [...relationEdges, ...capabilityParentEdges, ...capabilityDependencyEdges, ...realizationEdges];

      const layoutedNodes = calculateDagreLayout(nodes, edges, {
        direction: (currentView.layoutDirection as 'TB' | 'LR' | 'BT' | 'RL') || 'TB',
      });

      const componentPositions = layoutedNodes
        .filter((node) => node.type === 'component')
        .map((node) => ({
          componentId: node.id,
          x: node.position.x,
          y: node.position.y,
        }));

      const capabilityPositionUpdates = layoutedNodes
        .filter((node) => node.type === 'capability')
        .map((node) => ({
          capabilityId: node.id,
          x: node.position.x,
          y: node.position.y,
        }));

      await apiClient.updateMultiplePositions(currentView.id, { positions: componentPositions });

      for (const capPos of capabilityPositionUpdates) {
        await apiClient.updateCapabilityPositionInView(currentView.id, capPos.capabilityId, capPos.x, capPos.y);
      }

      const updatedView = await apiClient.getViewById(currentView.id);
      set({ currentView: updatedView });

      toast.success(`Layout applied to ${nodes.length} elements`);
    } catch (error) {
      const errorMessage = error instanceof ApiError
        ? error.message
        : 'Failed to apply layout';

      toast.error(errorMessage);
      throw error;
    }
  },
});
