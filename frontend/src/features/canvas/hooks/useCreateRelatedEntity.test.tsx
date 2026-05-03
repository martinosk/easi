import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook } from '@testing-library/react';
import React from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { ViewId } from '../../../api/types';
import { useAppStore } from '../../../store/appStore';
import type { RelatedLink } from '../../../utils/xRelated';
import type { RelationSubType } from '../utils/relationDispatch';
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

const vendorEntry: RelatedLink = {
  href: '/api/v1/components',
  methods: ['POST'],
  title: 'Component (purchased-from)',
  targetType: 'component',
  relationType: 'origin-purchased-from',
};

const teamEntry: RelatedLink = {
  href: '/api/v1/components',
  methods: ['POST'],
  title: 'Component (built-by)',
  targetType: 'component',
  relationType: 'origin-built-by',
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
      dynamicRelations: [],
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
      dynamicRelations: [],
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

interface StartArgs {
  entry: RelatedLink;
  sourceEntityId: string;
  newEntityId: string;
  side?: 'top' | 'right' | 'bottom' | 'left';
  sourcePosition?: { x: number; y: number };
  relationSubType?: RelationSubType;
}

async function runFlow(args: StartArgs) {
  const { result } = renderOrchestrator();
  act(() => {
    result.current.start({
      entry: args.entry,
      sourceEntityId: args.sourceEntityId,
      side: args.side ?? 'right',
      sourcePosition: args.sourcePosition ?? { x: 100, y: 200 },
      relationSubType: args.relationSubType,
    });
  });
  await act(async () => {
    await result.current.handleEntityCreated(args.newEntityId);
  });
  return result;
}

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
    expect(result.current.pending).toMatchObject({
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
        sourcePosition: { x: 0, y: 0 },
      });
    });
    act(() => {
      result.current.cancel();
    });
    expect(result.current.pending).toBeNull();
  });
});

describe('useCreateRelatedEntity — handleEntityCreated regular mode dispatch', () => {
  it.each<[string, StartArgs, () => void]>([
    [
      'component-relation defaults to Triggers when no subtype is supplied',
      { entry: componentEntry, sourceEntityId: 'comp-source', newEntityId: 'comp-new' },
      () =>
        expect(createRelationMutate).toHaveBeenCalledWith(
          expect.objectContaining({
            sourceComponentId: 'comp-source',
            targetComponentId: 'comp-new',
            relationType: 'Triggers',
          }),
        ),
    ],
    [
      'component-relation honors explicit Serves subtype',
      { entry: componentEntry, sourceEntityId: 'comp-source', newEntityId: 'comp-new', relationSubType: 'Serves' },
      () =>
        expect(createRelationMutate).toHaveBeenCalledWith(
          expect.objectContaining({ relationType: 'Serves' }),
        ),
    ],
    [
      'capability-parent makes the new capability the child of the source',
      { entry: capabilityParentEntry, sourceEntityId: 'cap-source', newEntityId: 'cap-new', side: 'bottom' },
      () =>
        expect(changeCapabilityParentMutate).toHaveBeenCalledWith(
          expect.objectContaining({ id: 'cap-new', newParentId: 'cap-source' }),
        ),
    ],
    [
      'capability-realization links source capability to new component',
      { entry: realizationEntry, sourceEntityId: 'cap-source', newEntityId: 'comp-new' },
      () =>
        expect(linkSystemToCapabilityMutate).toHaveBeenCalledWith(
          expect.objectContaining({
            capabilityId: 'cap-source',
            request: expect.objectContaining({ componentId: 'comp-new' }),
          }),
        ),
    ],
    [
      'origin-acquired-via attaches new component to source acquired entity',
      { entry: acquiredViaEntry, sourceEntityId: 'acq-1', newEntityId: 'comp-new' },
      () =>
        expect(linkComponentToAcquiredEntityMutate).toHaveBeenCalledWith(
          expect.objectContaining({ componentId: 'comp-new', entityId: 'acq-1' }),
        ),
    ],
    [
      'origin-purchased-from attaches new component to source vendor',
      { entry: vendorEntry, sourceEntityId: 'vendor-1', newEntityId: 'comp-new' },
      () =>
        expect(linkComponentToVendorMutate).toHaveBeenCalledWith(
          expect.objectContaining({ componentId: 'comp-new', vendorId: 'vendor-1' }),
        ),
    ],
    [
      'origin-built-by attaches new component to source internal team',
      { entry: teamEntry, sourceEntityId: 'team-1', newEntityId: 'comp-new' },
      () =>
        expect(linkComponentToInternalTeamMutate).toHaveBeenCalledWith(
          expect.objectContaining({ componentId: 'comp-new', teamId: 'team-1' }),
        ),
    ],
  ])('%s', async (_name, args, assertion) => {
    await runFlow(args);
    assertion();
  });
});

describe('useCreateRelatedEntity — handleEntityCreated regular mode add-to-view', () => {
  it('adds the new component to the view at the offset right of the source', async () => {
    await runFlow({ entry: componentEntry, sourceEntityId: 'comp-source', newEntityId: 'comp-new' });
    expect(addComponentToViewMutate).toHaveBeenCalledTimes(1);
    const call = addComponentToViewMutate.mock.calls[0][0];
    expect(call.viewId).toBe('v1');
    expect(call.request.componentId).toBe('comp-new');
    expect(call.request.x).toBeGreaterThan(100);
    expect(call.request.y).toBe(200);
  });

  it('clears pending after a successful create', async () => {
    const result = await runFlow({ entry: componentEntry, sourceEntityId: 'comp-source', newEntityId: 'comp-new' });
    expect(result.current.pending).toBeNull();
  });

  it('uses addCapabilityToView for capability-parent flows', async () => {
    await runFlow({ entry: capabilityParentEntry, sourceEntityId: 'cap-source', newEntityId: 'cap-new' });
    expect(addCapabilityToViewMutate).toHaveBeenCalled();
  });

  it('uses addOriginEntityToView for origin-acquired-via flows', async () => {
    await runFlow({
      entry: { ...acquiredViaEntry, targetType: 'acquiredEntity' },
      sourceEntityId: 'acq-source',
      newEntityId: 'acq-new',
    });
    expect(addOriginEntityToViewMutate).toHaveBeenCalled();
  });
});

describe('useCreateRelatedEntity — relation failure rollback', () => {
  it('keeps the just-created entity on the canvas by still adding it to the view (spec rule 13)', async () => {
    createRelationMutate.mockRejectedValueOnce(new Error('relation not allowed'));

    const result = await runFlow({ entry: componentEntry, sourceEntityId: 'comp-source', newEntityId: 'comp-new' });

    expect(toastError).toHaveBeenCalledWith(expect.stringMatching(/Component \(related\)/));
    expect(addComponentToViewMutate).toHaveBeenCalledTimes(1);
    expect(addComponentToViewMutate.mock.calls[0][0].request.componentId).toBe('comp-new');
    expect(result.current.pending).toBeNull();
  });
});

describe('useCreateRelatedEntity — dynamic mode', () => {
  it('writes the entity, position, and relation to the draft store and skips backend calls', async () => {
    enableDynamicMode();
    await runFlow({ entry: componentEntry, sourceEntityId: 'comp-source', newEntityId: 'comp-new' });

    expect(createRelationMutate).not.toHaveBeenCalled();
    expect(addComponentToViewMutate).not.toHaveBeenCalled();

    const state = useAppStore.getState();
    expect(state.dynamicEntities.map((e) => e.id)).toContain('comp-new');
    expect(state.dynamicPositions['comp-new']).toBeDefined();
    expect(state.dynamicPositions['comp-new'].x).toBeGreaterThan(100);
    expect(state.dynamicPositions['comp-new'].y).toBe(200);
    expect(state.dynamicRelations).toEqual([
      expect.objectContaining({
        kind: 'component-relation',
        sourceComponentId: 'comp-source',
        targetComponentId: 'comp-new',
        relationSubType: 'Triggers',
      }),
    ]);
  });

  it('records the chosen Serves subtype on the draft relation', async () => {
    enableDynamicMode();
    await runFlow({
      entry: componentEntry,
      sourceEntityId: 'comp-source',
      newEntityId: 'comp-new',
      relationSubType: 'Serves',
    });

    const state = useAppStore.getState();
    expect(state.dynamicRelations[0]).toMatchObject({ kind: 'component-relation', relationSubType: 'Serves' });
  });
});
