import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook } from '@testing-library/react';
import React from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { ViewId } from '../../../api/types';
import { useAppStore } from '../../../store/appStore';
import type { RelatedLink } from '../../../utils/xRelated';
import { useCreateRelatedEntity } from './useCreateRelatedEntity';

const toastError = vi.fn();
vi.mock('react-hot-toast', () => ({
  default: { error: (...args: unknown[]) => toastError(...args), success: vi.fn() },
}));

vi.mock('../../views/hooks/useCurrentView', () => ({
  useCurrentView: () => ({ currentView: { id: 'v1', _links: { edit: { href: '/x', method: 'PUT' } } }, currentViewId: 'v1' as ViewId, isLoading: false, error: null }),
}));

const createRelationMutate = vi.fn().mockResolvedValue({ id: 'rel-1' });
const changeCapabilityParentMutate = vi.fn().mockResolvedValue(undefined);
const linkSystemToCapabilityMutate = vi.fn().mockResolvedValue(undefined);
const linkComponentToAcquiredEntityMutate = vi.fn().mockResolvedValue(undefined);
const linkComponentToVendorMutate = vi.fn().mockResolvedValue(undefined);
const linkComponentToInternalTeamMutate = vi.fn().mockResolvedValue(undefined);
const addComponentToViewMutate = vi.fn().mockResolvedValue(undefined);
const addCapabilityToViewMutate = vi.fn().mockResolvedValue(undefined);
const addOriginEntityToViewMutate = vi.fn().mockResolvedValue(undefined);

vi.mock('../../relations/hooks/useRelations', () => ({
  useCreateRelation: () => ({ mutateAsync: createRelationMutate }),
}));

vi.mock('../../capabilities/hooks/useCapabilities', () => ({
  useChangeCapabilityParent: () => ({ mutateAsync: changeCapabilityParentMutate }),
  useLinkSystemToCapability: () => ({ mutateAsync: linkSystemToCapabilityMutate }),
}));

vi.mock('../../origin-entities/hooks', () => ({
  useLinkComponentToAcquiredEntity: () => ({ mutateAsync: linkComponentToAcquiredEntityMutate }),
  useLinkComponentToVendor: () => ({ mutateAsync: linkComponentToVendorMutate }),
  useLinkComponentToInternalTeam: () => ({ mutateAsync: linkComponentToInternalTeamMutate }),
}));

vi.mock('../../views/hooks/useViews', () => ({
  useAddComponentToView: () => ({ mutateAsync: addComponentToViewMutate }),
  useAddCapabilityToView: () => ({ mutateAsync: addCapabilityToViewMutate }),
  useAddOriginEntityToView: () => ({ mutateAsync: addOriginEntityToViewMutate }),
}));

const componentEntry: RelatedLink = {
  href: '/api/v1/components',
  methods: ['POST'],
  title: 'Component (related)',
  targetType: 'component',
  relationType: 'component-relation',
};

const capabilityParentEntry: RelatedLink = {
  href: '/api/v1/capabilities',
  methods: ['POST'],
  title: 'Capability (child of)',
  targetType: 'capability',
  relationType: 'capability-parent',
};

const realizationEntry: RelatedLink = {
  href: '/api/v1/components',
  methods: ['POST'],
  title: 'Component (realization)',
  targetType: 'component',
  relationType: 'capability-realization',
};

const acquiredViaEntry: RelatedLink = {
  href: '/api/v1/components',
  methods: ['POST'],
  title: 'Component (acquired-via)',
  targetType: 'component',
  relationType: 'origin-acquired-via',
};

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

function resetStore() {
  act(() => {
    useAppStore.setState({
      currentViewId: 'v1' as ViewId,
      dynamicViewId: null,
      dynamicEntities: [],
      dynamicOriginal: null,
      dynamicPositions: {},
      draftsByView: {},
    });
  });
}

function enableDynamicMode() {
  act(() => {
    useAppStore.setState({
      currentViewId: 'v1' as ViewId,
      dynamicViewId: 'v1',
      dynamicOriginal: { entities: [], positions: {} },
      dynamicEntities: [],
      dynamicPositions: {},
    });
  });
}

beforeEach(() => {
  vi.clearAllMocks();
  resetStore();
});

afterEach(() => {
  resetStore();
});

const renderOrchestrator = () =>
  renderHook(() => useCreateRelatedEntity(), { wrapper: createWrapper() });

describe('useCreateRelatedEntity — pending state', () => {
  it('starts with no pending creation', () => {
    const { result } = renderOrchestrator();
    expect(result.current.pending).toBeNull();
  });

  it('sets pending when start is called', () => {
    const { result } = renderOrchestrator();
    act(() => {
      result.current.start({
        entry: componentEntry,
        sourceEntityId: 'comp-a',
        side: 'right',
        sourcePosition: { x: 100, y: 200 },
      });
    });
    expect(result.current.pending).toEqual({
      entry: componentEntry,
      sourceEntityId: 'comp-a',
      side: 'right',
      sourcePosition: { x: 100, y: 200 },
    });
  });

  it('clears pending when cancel is called', () => {
    const { result } = renderOrchestrator();
    act(() => {
      result.current.start({
        entry: componentEntry,
        sourceEntityId: 'comp-a',
        side: 'right',
        sourcePosition: { x: 100, y: 200 },
      });
    });
    act(() => {
      result.current.cancel();
    });
    expect(result.current.pending).toBeNull();
  });
});

describe('useCreateRelatedEntity — handleEntityCreated (regular mode)', () => {
  it('creates a component-to-component relation with the source as origin', async () => {
    const { result } = renderOrchestrator();
    act(() => {
      result.current.start({
        entry: componentEntry,
        sourceEntityId: 'comp-source',
        side: 'right',
        sourcePosition: { x: 100, y: 200 },
      });
    });

    await act(async () => {
      await result.current.handleEntityCreated('comp-new');
    });

    expect(createRelationMutate).toHaveBeenCalledTimes(1);
    expect(createRelationMutate).toHaveBeenCalledWith(
      expect.objectContaining({
        sourceComponentId: 'comp-source',
        targetComponentId: 'comp-new',
      }),
    );
  });

  it('adds the new component to the view at the offset position to the right of the source', async () => {
    const { result } = renderOrchestrator();
    act(() => {
      result.current.start({
        entry: componentEntry,
        sourceEntityId: 'comp-source',
        side: 'right',
        sourcePosition: { x: 100, y: 200 },
      });
    });

    await act(async () => {
      await result.current.handleEntityCreated('comp-new');
    });

    expect(addComponentToViewMutate).toHaveBeenCalledTimes(1);
    const call = addComponentToViewMutate.mock.calls[0][0];
    expect(call.viewId).toBe('v1');
    expect(call.request.componentId).toBe('comp-new');
    expect(call.request.x).toBeGreaterThan(100);
    expect(call.request.y).toBe(200);
  });

  it('clears pending after a successful create', async () => {
    const { result } = renderOrchestrator();
    act(() => {
      result.current.start({
        entry: componentEntry,
        sourceEntityId: 'comp-source',
        side: 'right',
        sourcePosition: { x: 0, y: 0 },
      });
    });

    await act(async () => {
      await result.current.handleEntityCreated('comp-new');
    });

    expect(result.current.pending).toBeNull();
  });

  it('dispatches capability-parent so the new capability becomes the child', async () => {
    const { result } = renderOrchestrator();
    act(() => {
      result.current.start({
        entry: capabilityParentEntry,
        sourceEntityId: 'cap-source',
        side: 'bottom',
        sourcePosition: { x: 0, y: 0 },
      });
    });

    await act(async () => {
      await result.current.handleEntityCreated('cap-new');
    });

    expect(changeCapabilityParentMutate).toHaveBeenCalledWith(
      expect.objectContaining({ id: 'cap-new', newParentId: 'cap-source' }),
    );
    expect(addCapabilityToViewMutate).toHaveBeenCalled();
  });

  it('dispatches capability-realization with the source capability and new component', async () => {
    const { result } = renderOrchestrator();
    act(() => {
      result.current.start({
        entry: realizationEntry,
        sourceEntityId: 'cap-source',
        side: 'right',
        sourcePosition: { x: 0, y: 0 },
      });
    });

    await act(async () => {
      await result.current.handleEntityCreated('comp-new');
    });

    expect(linkSystemToCapabilityMutate).toHaveBeenCalledWith(
      expect.objectContaining({
        capabilityId: 'cap-source',
        request: expect.objectContaining({ componentId: 'comp-new' }),
      }),
    );
    expect(addComponentToViewMutate).toHaveBeenCalled();
  });

  it('dispatches origin-acquired-via with the new component and source acquired entity', async () => {
    const { result } = renderOrchestrator();
    act(() => {
      result.current.start({
        entry: acquiredViaEntry,
        sourceEntityId: 'acq-1',
        side: 'right',
        sourcePosition: { x: 0, y: 0 },
      });
    });

    await act(async () => {
      await result.current.handleEntityCreated('comp-new');
    });

    expect(linkComponentToAcquiredEntityMutate).toHaveBeenCalledWith(
      expect.objectContaining({ componentId: 'comp-new', entityId: 'acq-1' }),
    );
  });

  it('shows an actionable toast when the relation call fails and keeps no orphan rollback', async () => {
    createRelationMutate.mockRejectedValueOnce(new Error('relation not allowed'));

    const { result } = renderOrchestrator();
    act(() => {
      result.current.start({
        entry: componentEntry,
        sourceEntityId: 'comp-source',
        side: 'right',
        sourcePosition: { x: 0, y: 0 },
      });
    });

    await act(async () => {
      await result.current.handleEntityCreated('comp-new');
    });

    expect(toastError).toHaveBeenCalledWith(
      expect.stringMatching(/Component \(related\)/),
    );
    expect(addComponentToViewMutate).not.toHaveBeenCalled();
    expect(result.current.pending).toBeNull();
  });
});

describe('useCreateRelatedEntity — dynamic mode', () => {
  it('writes the new entity, relation-implied membership, and position to the draft store without backend calls', async () => {
    enableDynamicMode();
    const { result } = renderOrchestrator();

    act(() => {
      result.current.start({
        entry: componentEntry,
        sourceEntityId: 'comp-source',
        side: 'right',
        sourcePosition: { x: 100, y: 200 },
      });
    });

    await act(async () => {
      await result.current.handleEntityCreated('comp-new');
    });

    expect(createRelationMutate).not.toHaveBeenCalled();
    expect(addComponentToViewMutate).not.toHaveBeenCalled();

    const state = useAppStore.getState();
    const ids = state.dynamicEntities.map((e) => e.id);
    expect(ids).toContain('comp-new');
    expect(state.dynamicPositions['comp-new']).toBeDefined();
    expect(state.dynamicPositions['comp-new'].x).toBeGreaterThan(100);
    expect(state.dynamicPositions['comp-new'].y).toBe(200);
  });
});
