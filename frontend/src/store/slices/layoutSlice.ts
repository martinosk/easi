import type { StateCreator } from 'zustand';
import type { Component, Relation, View, Capability, CapabilityDependency, CapabilityRealization, ViewComponent } from '../../api/types';
import type { ComponentId, CapabilityId, Position, EdgeType, LayoutDirection } from '../types/storeTypes';
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

type CanvasCapability = { capabilityId: string; x: number; y: number };

type StoreWithDependencies = {
  currentView: View | null;
  components: Component[];
  relations: Relation[];
  capabilities: Capability[];
  capabilityDependencies: CapabilityDependency[];
  capabilityRealizations: CapabilityRealization[];
  canvasCapabilities: CanvasCapability[];
};

interface LayoutNode {
  id: string;
  type?: string;
  position: Position;
  data: { label: string; description?: string };
}

interface LayoutEdge {
  id: string;
  source: string;
  target: string;
}

function buildComponentNodes(
  components: Component[],
  viewComponents: ViewComponent[]
): LayoutNode[] {
  return components
    .filter((component) =>
      viewComponents.some((vc) => vc.componentId === component.id)
    )
    .map((component) => {
      const viewComponent = viewComponents.find(
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
}

function buildCapabilityNodes(
  canvasCapabilities: CanvasCapability[],
  capabilities: Capability[]
): LayoutNode[] {
  return canvasCapabilities
    .map((canvasCapability) => {
      const capability = capabilities.find((c) => c.id === canvasCapability.capabilityId);
      if (!capability) return null;

      return {
        id: capability.id,
        type: 'capability' as const,
        position: { x: canvasCapability.x, y: canvasCapability.y },
        data: {
          label: capability.name,
          description: capability.description,
        },
      } satisfies LayoutNode;
    })
    .filter((node): node is NonNullable<typeof node> => node !== null);
}

function buildRelationEdges(
  relations: Relation[],
  viewComponents: ViewComponent[]
): LayoutEdge[] {
  return relations
    .filter((relation) => {
      const sourceInView = viewComponents.some(
        (vc) => vc.componentId === relation.sourceComponentId
      );
      const targetInView = viewComponents.some(
        (vc) => vc.componentId === relation.targetComponentId
      );
      return sourceInView && targetInView;
    })
    .map((relation) => ({
      id: relation.id,
      source: relation.sourceComponentId,
      target: relation.targetComponentId,
    }));
}

function buildCapabilityParentEdges(
  canvasCapabilities: CanvasCapability[],
  capabilities: Capability[]
): LayoutEdge[] {
  return canvasCapabilities
    .map((canvasCapability) => {
      const capability = capabilities.find((c) => c.id === canvasCapability.capabilityId);
      if (!capability || !capability.parentId) return null;

      const parentInView = canvasCapabilities.some((cc) => cc.capabilityId === capability.parentId);
      if (!parentInView) return null;

      return {
        id: `parent-${capability.id}`,
        source: capability.parentId as string,
        target: capability.id as string,
      };
    })
    .filter((edge): edge is LayoutEdge => edge !== null);
}

function buildCapabilityDependencyEdges(
  capabilityDependencies: CapabilityDependency[],
  canvasCapabilities: CanvasCapability[]
): LayoutEdge[] {
  return capabilityDependencies
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
}

function buildRealizationEdges(
  capabilityRealizations: CapabilityRealization[],
  canvasCapabilities: CanvasCapability[],
  viewComponents: ViewComponent[]
): LayoutEdge[] {
  return capabilityRealizations
    .filter((real) => {
      const capabilityInView = canvasCapabilities.some((cc) => cc.capabilityId === real.capabilityId);
      const componentInView = viewComponents.some((vc) => vc.componentId === real.componentId);
      return capabilityInView && componentInView;
    })
    .map((real) => ({
      id: real.id,
      source: real.componentId,
      target: real.capabilityId,
    }));
}

function extractComponentPositions(layoutedNodes: LayoutNode[]): Array<{ componentId: ComponentId; x: number; y: number }> {
  return layoutedNodes
    .filter((node) => node.type === 'component')
    .map((node) => ({
      componentId: node.id as ComponentId,
      x: node.position.x,
      y: node.position.y,
    }));
}

function extractCapabilityPositions(layoutedNodes: LayoutNode[]): Array<{ capabilityId: CapabilityId; x: number; y: number }> {
  return layoutedNodes
    .filter((node) => node.type === 'capability')
    .map((node) => ({
      capabilityId: node.id as CapabilityId,
      x: node.position.x,
      y: node.position.y,
    }));
}

type ViewPropertyKey = 'edgeType' | 'layoutDirection' | 'colorScheme';

async function updateViewProperty<K extends ViewPropertyKey>(
  currentView: View,
  propertyKey: K,
  newValue: string,
  set: (partial: { currentView: View }) => void,
  apiCall: () => Promise<unknown>,
  successMessage: string,
  errorMessage: string
): Promise<void> {
  const previousValue = currentView[propertyKey];

  set({
    currentView: {
      ...currentView,
      [propertyKey]: newValue,
    },
  });

  await optimisticUpdate(
    apiCall,
    () => {},
    () => set({
      currentView: {
        ...currentView,
        [propertyKey]: previousValue,
      },
    }),
    successMessage,
    errorMessage
  );
}

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
    if (!currentView) return;

    await updateViewProperty(
      currentView,
      'edgeType',
      edgeType,
      set,
      () => apiClient.updateViewEdgeType(currentView.id, { edgeType }),
      'Edge type updated',
      'Failed to update edge type'
    );
  },

  setLayoutDirection: async (layoutDirection: LayoutDirection) => {
    const { currentView } = get();
    if (!currentView) return;

    await updateViewProperty(
      currentView,
      'layoutDirection',
      layoutDirection,
      set,
      () => apiClient.updateViewLayoutDirection(currentView.id, { layoutDirection }),
      'Layout direction updated',
      'Failed to update layout direction'
    );
  },

  setColorScheme: async (colorScheme: string) => {
    const { currentView } = get();
    if (!currentView) return;

    await updateViewProperty(
      currentView,
      'colorScheme',
      colorScheme,
      set,
      () => apiClient.updateViewColorScheme(currentView.id, { colorScheme }),
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

      const componentNodes = buildComponentNodes(components, currentView.components);
      const capabilityNodes = buildCapabilityNodes(canvasCapabilities, capabilities);
      const nodes = [...componentNodes, ...capabilityNodes];

      const relationEdges = buildRelationEdges(relations, currentView.components);
      const capabilityParentEdges = buildCapabilityParentEdges(canvasCapabilities, capabilities);
      const capabilityDependencyEdges = buildCapabilityDependencyEdges(capabilityDependencies, canvasCapabilities);
      const realizationEdges = buildRealizationEdges(capabilityRealizations, canvasCapabilities, currentView.components);
      const edges = [...relationEdges, ...capabilityParentEdges, ...capabilityDependencyEdges, ...realizationEdges];

      const layoutedNodes = calculateDagreLayout(nodes, edges, {
        direction: (currentView.layoutDirection as 'TB' | 'LR' | 'BT' | 'RL') || 'TB',
      }) as LayoutNode[];

      const componentPositions = extractComponentPositions(layoutedNodes);
      const capabilityPositionUpdates = extractCapabilityPositions(layoutedNodes);

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
