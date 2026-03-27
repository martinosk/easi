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

function toPositionMap(elements: { elementId: string; x: number; y: number }[]): PositionMap {
  const map: PositionMap = {};
  for (const elem of elements) {
    map[elem.elementId] = { x: elem.x, y: elem.y };
  }
  return map;
}

function rollbackPosition(prev: PositionMap, elementId: string, previous: Position | undefined): PositionMap {
  if (previous) return { ...prev, [elementId]: previous };
  const { [elementId]: _, ...rest } = prev;
  return rest;
}

function applyBatch(prev: PositionMap, updates: BatchUpdateItem[]): PositionMap {
  const next = { ...prev };
  for (const u of updates) {
    next[u.elementId] = { x: u.x, y: u.y };
  }
  return next;
}

function useLayoutInitializer(contextType: LayoutContextType, contextRef: string | null) {
  const [layout, setLayout] = useState<LayoutContainer | null>(null);
  const [positions, setPositions] = useState<PositionMap>({});
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const initializeLayout = useCallback(async () => {
    if (!contextRef) {
      setLayout(null);
      setPositions({});
      return;
    }
    setIsLoading(true);
    setError(null);
    try {
      const container =
        (await apiClient.getLayout(contextType, contextRef)) ??
        (await apiClient.upsertLayout(contextType, contextRef, {}));
      setLayout(container);
      setPositions(toPositionMap(container.elements || []));
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

  return { layout, setLayout, positions, setPositions, isLoading, error, initializeLayout };
}

function useElementPositionUpdater(
  contextType: LayoutContextType, contextRef: string | null,
  layout: LayoutContainer | null, setPositions: React.Dispatch<React.SetStateAction<PositionMap>>
) {
  return useCallback(
    async (elementId: string, x: number, y: number, options?: Partial<ElementPositionInput>) => {
      if (!contextRef || !layout) return;
      let previousPosition: Position | undefined;
      setPositions((prev) => {
        previousPosition = prev[elementId];
        return { ...prev, [elementId]: { x, y } };
      });
      try {
        await apiClient.upsertElementPosition(contextType, contextRef, elementId, { x, y, ...options });
      } catch (err) {
        setPositions((prev) => rollbackPosition(prev, elementId, previousPosition));
        throw err;
      }
    },
    [contextType, contextRef, layout, setPositions]
  );
}

function useBatchPositionUpdater(
  contextType: LayoutContextType, contextRef: string | null,
  layout: LayoutContainer | null, setPositions: React.Dispatch<React.SetStateAction<PositionMap>>
) {
  return useCallback(
    async (updates: BatchUpdateItem[]) => {
      if (!contextRef || !layout) return;
      let snapshot: PositionMap = {};
      setPositions((prev) => {
        snapshot = { ...prev };
        return applyBatch(prev, updates);
      });
      try {
        await apiClient.batchUpdateElements(contextType, contextRef, updates);
      } catch (err) {
        setPositions(snapshot);
        throw err;
      }
    },
    [contextType, contextRef, layout, setPositions]
  );
}

export function useLayout(
  contextType: LayoutContextType,
  contextRef: string | null
): UseLayoutResult {
  const { layout, setLayout, positions, setPositions, isLoading, error, initializeLayout } =
    useLayoutInitializer(contextType, contextRef);
  const preferences = useMemo(() => layout?.preferences ?? {}, [layout?.preferences]);
  const updateElementPosition = useElementPositionUpdater(contextType, contextRef, layout, setPositions);
  const batchUpdatePositions = useBatchPositionUpdater(contextType, contextRef, layout, setPositions);

  const updatePreferences = useCallback(
    async (newPreferences: Record<string, unknown>) => {
      if (!contextRef || !layout) return;
      const previousLayout = layout;
      setLayout((prev) =>
        prev ? { ...prev, preferences: { ...prev.preferences, ...newPreferences } } : prev
      );
      try {
        await apiClient.updateLayoutPreferences(contextType, contextRef, newPreferences, layout.version);
      } catch (err) {
        setLayout(previousLayout);
        throw err;
      }
    },
    [contextType, contextRef, layout, setLayout]
  );

  return useMemo(
    () => ({
      layout, positions, preferences, isLoading, error,
      updateElementPosition, batchUpdatePositions, updatePreferences,
      refetch: initializeLayout,
    }),
    [layout, positions, preferences, isLoading, error,
     updateElementPosition, batchUpdatePositions, updatePreferences, initializeLayout]
  );
}
