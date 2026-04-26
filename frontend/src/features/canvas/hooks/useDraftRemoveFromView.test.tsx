import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook } from '@testing-library/react';
import React from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type {
  Capability,
  CapabilityId,
  CapabilityRealization,
  ComponentId,
  OriginRelationship,
  Relation,
  ViewId,
} from '../../../api/types';
import { useAppStore } from '../../../store/appStore';
import { useDraftRemoveFromView } from './useDraftRemoveFromView';

vi.mock('../../views/hooks/useCurrentView', () => ({
  useCurrentView: () => ({ currentView: null, currentViewId: 'v1' as ViewId, isLoading: false, error: null }),
}));

const mockData = {
  relations: [] as unknown as Relation[],
  capabilities: [] as unknown as Capability[],
  realizations: [] as unknown as CapabilityRealization[],
  originRelationships: [] as OriginRelationship[],
};

vi.mock('../../relations/hooks/useRelations', () => ({
  useRelations: () => ({ data: mockData.relations }),
}));

vi.mock('../../capabilities/hooks/useCapabilities', () => ({
  useCapabilities: () => ({ data: mockData.capabilities }),
  useRealizations: () => ({ data: mockData.realizations }),
}));

vi.mock('../../origin-entities/hooks/useOriginRelationships', () => ({
  useOriginRelationshipsQuery: () => ({ data: mockData.originRelationships }),
}));

function resetMockData() {
  mockData.relations = [];
  mockData.capabilities = [];
  mockData.realizations = [];
  mockData.originRelationships = [];
}

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

function seedDraft(entities: { id: string; type: 'component' | 'capability' | 'originEntity' }[]) {
  act(() => {
    useAppStore.setState({
      currentViewId: 'v1' as ViewId,
      dynamicViewId: 'v1' as ViewId,
      dynamicEntities: entities,
      dynamicOriginal: { entities: [...entities], positions: {} },
      dynamicPositions: {},
    });
  });
}

function renderRemove() {
  return renderHook(() => useDraftRemoveFromView(), { wrapper: createWrapper() });
}

describe('useDraftRemoveFromView', () => {
  beforeEach(() => {
    resetMockData();
    act(() => {
      useAppStore.setState({
        currentViewId: null,
        dynamicViewId: null,
        dynamicEntities: [],
        dynamicOriginal: null,
        dynamicPositions: {},
        draftsByView: {},
      });
    });
  });

  afterEach(() => {
    resetMockData();
  });

  it('removes only the clicked entity, never related entities', () => {
    seedDraft([
      { id: 'comp-a', type: 'component' },
      { id: 'comp-b', type: 'component' },
      { id: 'comp-c', type: 'component' },
    ]);
    mockData.relations = [
      {
        id: 'rel-1',
        sourceComponentId: 'comp-a' as ComponentId,
        targetComponentId: 'comp-b' as ComponentId,
      } as unknown as Relation,
      {
        id: 'rel-2',
        sourceComponentId: 'comp-b' as ComponentId,
        targetComponentId: 'comp-c' as ComponentId,
      } as unknown as Relation,
    ];

    const { result } = renderRemove();

    act(() => {
      result.current('comp-b', 'component');
    });

    const remaining = useAppStore.getState().dynamicEntities.map((e) => e.id);
    expect(remaining.sort()).toEqual(['comp-a', 'comp-c']);
  });

  it('does not cascade-remove orphaned descendants', () => {
    seedDraft([
      { id: 'parent-cap', type: 'capability' },
      { id: 'child-cap-1', type: 'capability' },
      { id: 'child-cap-2', type: 'capability' },
      { id: 'child-cap-3', type: 'capability' },
      { id: 'child-cap-4', type: 'capability' },
      { id: 'child-cap-5', type: 'capability' },
    ]);
    mockData.capabilities = [
      { id: 'parent-cap' as unknown as CapabilityId, parentId: null } as unknown as Capability,
      {
        id: 'child-cap-1' as unknown as CapabilityId,
        parentId: 'parent-cap' as unknown as CapabilityId,
      } as unknown as Capability,
      {
        id: 'child-cap-2' as unknown as CapabilityId,
        parentId: 'parent-cap' as unknown as CapabilityId,
      } as unknown as Capability,
      {
        id: 'child-cap-3' as unknown as CapabilityId,
        parentId: 'parent-cap' as unknown as CapabilityId,
      } as unknown as Capability,
      {
        id: 'child-cap-4' as unknown as CapabilityId,
        parentId: 'parent-cap' as unknown as CapabilityId,
      } as unknown as Capability,
      {
        id: 'child-cap-5' as unknown as CapabilityId,
        parentId: 'parent-cap' as unknown as CapabilityId,
      } as unknown as Capability,
    ];

    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true);

    const { result } = renderRemove();

    act(() => {
      result.current('parent-cap', 'capability');
    });

    const remaining = useAppStore
      .getState()
      .dynamicEntities.map((e) => e.id)
      .sort();
    expect(remaining).toEqual(['child-cap-1', 'child-cap-2', 'child-cap-3', 'child-cap-4', 'child-cap-5']);
    expect(confirmSpy).not.toHaveBeenCalled();

    confirmSpy.mockRestore();
  });

  it('returns false and does nothing when no draft is active for the current view', () => {
    act(() => {
      useAppStore.setState({
        currentViewId: 'v1' as ViewId,
        dynamicViewId: null,
        dynamicEntities: [{ id: 'comp-a', type: 'component' }],
        dynamicOriginal: null,
      });
    });

    const { result } = renderRemove();

    let returnValue: boolean | undefined;
    act(() => {
      returnValue = result.current('comp-a', 'component');
    });

    expect(returnValue).toBe(false);
    expect(useAppStore.getState().dynamicEntities).toEqual([{ id: 'comp-a', type: 'component' }]);
  });
});
