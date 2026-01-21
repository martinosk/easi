import { useCallback, useMemo } from 'react';
import { useLayout } from '../../../hooks/useLayout';
import type { ViewId, ComponentId, CapabilityId, Position } from '../../../api/types';
import type { BatchUpdateItem } from '../../../api/types';

export interface CanvasPositionMap {
  [elementId: string]: Position;
}

export interface UseCanvasLayoutResult {
  positions: CanvasPositionMap;
  isLoading: boolean;
  error: Error | null;
  updateComponentPosition: (componentId: ComponentId, x: number, y: number) => Promise<void>;
  updateCapabilityPosition: (capabilityId: CapabilityId, x: number, y: number) => Promise<void>;
  updateOriginEntityPosition: (nodeId: string, x: number, y: number) => Promise<void>;
  batchUpdatePositions: (updates: BatchUpdateItem[]) => Promise<void>;
  getPositionForElement: (elementId: string) => Position | null;
  refetch: () => Promise<void>;
}

export function useCanvasLayout(viewId: ViewId | null): UseCanvasLayoutResult {
  const {
    positions,
    isLoading,
    error,
    updateElementPosition,
    batchUpdatePositions,
    refetch,
  } = useLayout('architecture-canvas', viewId);

  const updateComponentPosition = useCallback(
    async (componentId: ComponentId, x: number, y: number) => {
      await updateElementPosition(componentId, x, y);
    },
    [updateElementPosition]
  );

  const updateCapabilityPosition = useCallback(
    async (capabilityId: CapabilityId, x: number, y: number) => {
      await updateElementPosition(capabilityId, x, y);
    },
    [updateElementPosition]
  );

  const updateOriginEntityPosition = useCallback(
    async (nodeId: string, x: number, y: number) => {
      await updateElementPosition(nodeId, x, y);
    },
    [updateElementPosition]
  );

  const getPositionForElement = useCallback(
    (elementId: string): Position | null => {
      return positions[elementId] || null;
    },
    [positions]
  );

  return useMemo(
    () => ({
      positions,
      isLoading,
      error,
      updateComponentPosition,
      updateCapabilityPosition,
      updateOriginEntityPosition,
      batchUpdatePositions,
      getPositionForElement,
      refetch,
    }),
    [
      positions,
      isLoading,
      error,
      updateComponentPosition,
      updateCapabilityPosition,
      updateOriginEntityPosition,
      batchUpdatePositions,
      getPositionForElement,
      refetch,
    ]
  );
}
