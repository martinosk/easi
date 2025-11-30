import { useState, useEffect, useCallback, useMemo } from 'react';
import { apiClient } from '../../../api/client';
import type { BusinessDomainId, CapabilityId, ViewId, Position, View } from '../../../api/types';

export interface PositionMap {
  [capabilityId: string]: Position;
}

export interface UseGridPositionsResult {
  viewId: ViewId | null;
  positions: PositionMap;
  isLoading: boolean;
  error: Error | null;
  updatePosition: (capabilityId: CapabilityId, x: number, y: number) => Promise<void>;
  getPositionForCapability: (capabilityId: CapabilityId) => Position | null;
  refetch: () => Promise<void>;
}

function findViewForDomain(views: View[], domainId: BusinessDomainId): View | undefined {
  const viewName = `${domainId} Domain Layout`;
  return views.find((v) => v.name === viewName);
}

export function useGridPositions(domainId: BusinessDomainId | null): UseGridPositionsResult {
  const [viewId, setViewId] = useState<ViewId | null>(null);
  const [positions, setPositions] = useState<PositionMap>({});
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const initializeView = useCallback(async () => {
    if (!domainId) {
      setViewId(null);
      setPositions({});
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const views = await apiClient.getViews();
      let view = findViewForDomain(views, domainId);

      if (!view) {
        view = await apiClient.createView({
          name: `${domainId} Domain Layout`,
          description: `Grid layout for business domain ${domainId}`,
        });
      }

      setViewId(view.id);

      const posMap: PositionMap = {};
      for (const cap of view.capabilities || []) {
        posMap[cap.capabilityId] = { x: cap.x, y: cap.y };
      }
      setPositions(posMap);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to initialize view'));
    } finally {
      setIsLoading(false);
    }
  }, [domainId]);

  useEffect(() => {
    initializeView();
  }, [initializeView]);

  const updatePosition = useCallback(
    async (capabilityId: CapabilityId, x: number, y: number) => {
      if (!viewId) return;

      const existingPosition = positions[capabilityId];

      setPositions((prev) => ({
        ...prev,
        [capabilityId]: { x, y },
      }));

      try {
        if (existingPosition) {
          await apiClient.updateCapabilityPositionInView(viewId, capabilityId, { x, y });
        } else {
          await apiClient.addCapabilityToView(viewId, { capabilityId, x, y });
        }
      } catch (err) {
        if (existingPosition) {
          setPositions((prev) => ({
            ...prev,
            [capabilityId]: existingPosition,
          }));
        } else {
          setPositions((prev) => {
            const { [capabilityId]: _, ...rest } = prev;
            return rest;
          });
        }
        throw err;
      }
    },
    [viewId, positions]
  );

  const getPositionForCapability = useCallback(
    (capabilityId: CapabilityId): Position | null => {
      return positions[capabilityId] || null;
    },
    [positions]
  );

  return useMemo(
    () => ({
      viewId,
      positions,
      isLoading,
      error,
      updatePosition,
      getPositionForCapability,
      refetch: initializeView,
    }),
    [viewId, positions, isLoading, error, updatePosition, getPositionForCapability, initializeView]
  );
}
