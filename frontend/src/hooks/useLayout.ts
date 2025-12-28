import { useState, useEffect, useCallback, useMemo } from 'react';
import { apiClient } from '../api/client';
import type {
  LayoutContainer,
  LayoutContextType,
  ElementPositionInput,
  BatchUpdateItem,
  Position,
} from '../api/types';

export interface PositionMap {
  [elementId: string]: Position;
}

export interface UseLayoutResult {
  layout: LayoutContainer | null;
  positions: PositionMap;
  preferences: Record<string, unknown>;
  isLoading: boolean;
  error: Error | null;
  updateElementPosition: (
    elementId: string,
    x: number,
    y: number,
    options?: Partial<ElementPositionInput>
  ) => Promise<void>;
  batchUpdatePositions: (updates: BatchUpdateItem[]) => Promise<void>;
  updatePreferences: (preferences: Record<string, unknown>) => Promise<void>;
  refetch: () => Promise<void>;
}

export function useLayout(
  contextType: LayoutContextType,
  contextRef: string | null
): UseLayoutResult {
  const [layout, setLayout] = useState<LayoutContainer | null>(null);
  const [positions, setPositions] = useState<PositionMap>({});
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const preferences = useMemo(() => {
    return layout?.preferences ?? {};
  }, [layout?.preferences]);

  const initializeLayout = useCallback(async () => {
    if (!contextRef) {
      setLayout(null);
      setPositions({});
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      let container = await apiClient.getLayout(contextType, contextRef);

      if (!container) {
        container = await apiClient.upsertLayout(contextType, contextRef, {});
      }

      setLayout(container);

      const posMap: PositionMap = {};
      for (const elem of container.elements || []) {
        posMap[elem.elementId] = { x: elem.x, y: elem.y };
      }
      setPositions(posMap);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to initialize layout'));
    } finally {
      setIsLoading(false);
    }
  }, [contextType, contextRef]);

  useEffect(() => {
    initializeLayout();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [contextType, contextRef]);

  const updateElementPosition = useCallback(
    async (
      elementId: string,
      x: number,
      y: number,
      options?: Partial<ElementPositionInput>
    ) => {
      if (!contextRef || !layout) return;

      let previousPosition: Position | undefined;
      setPositions((prev) => {
        previousPosition = prev[elementId];
        return {
          ...prev,
          [elementId]: { x, y },
        };
      });

      try {
        await apiClient.upsertElementPosition(contextType, contextRef, elementId, {
          x,
          y,
          ...options,
        });
      } catch (err) {
        setPositions((prev) => {
          if (previousPosition) {
            return { ...prev, [elementId]: previousPosition };
          }
          const rest = Object.fromEntries(Object.entries(prev).filter(([key]) => key !== elementId));
          return rest;
        });
        throw err;
      }
    },
    [contextType, contextRef, layout]
  );

  const batchUpdatePositions = useCallback(
    async (updates: BatchUpdateItem[]) => {
      if (!contextRef || !layout) return;

      let previousPositions: PositionMap = {};
      setPositions((prev) => {
        previousPositions = { ...prev };
        const next = { ...prev };
        for (const update of updates) {
          next[update.elementId] = { x: update.x, y: update.y };
        }
        return next;
      });

      try {
        await apiClient.batchUpdateElements(contextType, contextRef, updates);
      } catch (err) {
        setPositions(previousPositions);
        throw err;
      }
    },
    [contextType, contextRef, layout]
  );

  const updatePreferences = useCallback(
    async (newPreferences: Record<string, unknown>) => {
      if (!contextRef || !layout) return;

      const previousLayout = layout;

      setLayout((prev) =>
        prev
          ? {
              ...prev,
              preferences: { ...prev.preferences, ...newPreferences },
            }
          : prev
      );

      try {
        await apiClient.updateLayoutPreferences(
          contextType,
          contextRef,
          newPreferences,
          layout.version
        );
      } catch (err) {
        setLayout(previousLayout);
        throw err;
      }
    },
    [contextType, contextRef, layout]
  );

  return useMemo(
    () => ({
      layout,
      positions,
      preferences,
      isLoading,
      error,
      updateElementPosition,
      batchUpdatePositions,
      updatePreferences,
      refetch: initializeLayout,
    }),
    [
      layout,
      positions,
      preferences,
      isLoading,
      error,
      updateElementPosition,
      batchUpdatePositions,
      updatePreferences,
      initializeLayout,
    ]
  );
}
