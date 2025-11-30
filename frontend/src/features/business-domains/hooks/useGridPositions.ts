import { useCallback, useMemo } from 'react';
import { useLayout } from '../../../hooks/useLayout';
import type { BusinessDomainId, CapabilityId, Position } from '../../../api/types';

export interface PositionMap {
  [capabilityId: string]: Position;
}

export interface UseGridPositionsResult {
  positions: PositionMap;
  isLoading: boolean;
  error: Error | null;
  updatePosition: (capabilityId: CapabilityId, x: number, y: number) => Promise<void>;
  getPositionForCapability: (capabilityId: CapabilityId) => Position | null;
  refetch: () => Promise<void>;
}

export function useGridPositions(domainId: BusinessDomainId | null): UseGridPositionsResult {
  const {
    positions,
    isLoading,
    error,
    updateElementPosition,
    refetch,
  } = useLayout('business-domain-grid', domainId);

  const updatePosition = useCallback(
    async (capabilityId: CapabilityId, x: number, y: number) => {
      await updateElementPosition(capabilityId, x, y);
    },
    [updateElementPosition]
  );

  const getPositionForCapability = useCallback(
    (capabilityId: CapabilityId): Position | null => {
      return positions[capabilityId] || null;
    },
    [positions]
  );

  return useMemo(
    () => ({
      positions,
      isLoading,
      error,
      updatePosition,
      getPositionForCapability,
      refetch,
    }),
    [positions, isLoading, error, updatePosition, getPositionForCapability, refetch]
  );
}
