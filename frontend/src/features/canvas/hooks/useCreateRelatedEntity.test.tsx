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

const triggersEntry: RelatedLink = {
  href: '/api/v1/components',
  methods: ['POST'],
  title: 'Component (triggers)',
  targetType: 'component',
  relationType: 'component-triggers',
};

const servesEntry: RelatedLink = {
  href: '/api/v1/components',
  methods: ['POST'],
  title: 'Component (serves)',
  targetType: 'component',
  relationType: 'component-serves',
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

const acquiredViaFromAcquiredEntity: RelatedLink = {
  href: '/api/v1/components',
  methods: ['POST'],
  title: 'Component (acquired-via)',
  targetType: 'component',
  relationType: 'origin-acquired-via',
};

const acquiredViaFromComponent: RelatedLink = {
  href: '/api/v1/acquired-entities',
  methods: ['POST'],
  title: 'Acquired Entity (acquired-via)',
  targetType: 'acquiredEntity',
  relationType: 'origin-acquired-via',
};

const purchasedFromVendor: RelatedLink = {
  href: '/api/v1/components',
  methods: ['POST'],
  title: 'Component (purchased-from)',
  targetType: 'component',
  relationType: 'origin-purchased-from',
};

const purchasedFromComponent: RelatedLink = {
  href: '/api/v1/vendors',
  methods: ['POST'],
  title: 'Vendor (purchased-from)',
  targetType: 'vendor',
  relationType: 'origin-purchased-from',
};

const builtByTeam: RelatedLink = {
  href: '/api/v1/components',
  methods: ['POST'],
  title: 'Component (built-by)',
  targetType: 'component',
  relationType: 'origin-built-by',
};

const builtByFromComponent: RelatedLink = {
  href: '/api/v1/internal-teams',
  methods: ['POST'],
  title: 'Internal Team (built-by)',
  targetType: 'internalTeam',
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

interface StartArgs {
  entry: RelatedLink;
  sourceEntityId: string;
  newEntityId: string;
  side?: 'top' | 'right' | 'bottom' | 'left';
  sourcePosition?: { x: number; y: number };
}

async function runFlow(args: StartArgs) {
  const { result } = renderOrchestrator();
  act(() => {
    result.current.start({
      entry: args.entry,
      sourceEntityId: args.sourceEntityId,
      side: args.side ?? 'right',
      sourcePosition: args.sourcePosition ?? { x: 100, y: 200 },
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
        entry: triggersEntry,
        sourceEntityId: 'comp-a',
        side: 'right',
        sourcePosition: { x: 100, y: 200 },
      });
    });
    expect(result.current.pending).toMatchObject({
      entry: triggersEntry,
      sourceEntityId: 'comp-a',
      side: 'right',
      sourcePosition: { x: 100, y: 200 },
    });
  });

  it('clears pending when cancel is called', () => {
    const { result } = renderOrchestrator();
    act(() => {
      result.current.start({
        entry: triggersEntry,
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
      'component-triggers dispatches a Triggers component-relation',
      { entry: triggersEntry, sourceEntityId: 'comp-source', newEntityId: 'comp-new' },
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
      'component-serves dispatches a Serves component-relation',
      { entry: servesEntry, sourceEntityId: 'comp-source', newEntityId: 'comp-new' },
      () =>
        expect(createRelationMutate).toHaveBeenCalledWith(
          expect.objectContaining({
            sourceComponentId: 'comp-source',
            targetComponentId: 'comp-new',
            relationType: 'Serves',
          }),
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
      'origin-acquired-via from acquired-entity source attaches new component',
      { entry: acquiredViaFromAcquiredEntity, sourceEntityId: 'acq-1', newEntityId: 'comp-new' },
      () =>
        expect(linkComponentToAcquiredEntityMutate).toHaveBeenCalledWith(
          expect.objectContaining({ componentId: 'comp-new', entityId: 'acq-1' }),
        ),
    ],
    [
      'origin-acquired-via from component source attaches new acquired-entity',
      { entry: acquiredViaFromComponent, sourceEntityId: 'comp-source', newEntityId: 'acq-new' },
      () =>
        expect(linkComponentToAcquiredEntityMutate).toHaveBeenCalledWith(
          expect.objectContaining({ componentId: 'comp-source', entityId: 'acq-new' }),
        ),
    ],
    [
      'origin-purchased-from from vendor source attaches new component',
      { entry: purchasedFromVendor, sourceEntityId: 'vendor-1', newEntityId: 'comp-new' },
      () =>
        expect(linkComponentToVendorMutate).toHaveBeenCalledWith(
          expect.objectContaining({ componentId: 'comp-new', vendorId: 'vendor-1' }),
        ),
    ],
    [
      'origin-purchased-from from component source attaches new vendor',
      { entry: purchasedFromComponent, sourceEntityId: 'comp-source', newEntityId: 'vendor-new' },
      () =>
        expect(linkComponentToVendorMutate).toHaveBeenCalledWith(
          expect.objectContaining({ componentId: 'comp-source', vendorId: 'vendor-new' }),
        ),
    ],
    [
      'origin-built-by from team source attaches new component',
      { entry: builtByTeam, sourceEntityId: 'team-1', newEntityId: 'comp-new' },
      () =>
        expect(linkComponentToInternalTeamMutate).toHaveBeenCalledWith(
          expect.objectContaining({ componentId: 'comp-new', teamId: 'team-1' }),
        ),
    ],
    [
      'origin-built-by from component source attaches new team',
      { entry: builtByFromComponent, sourceEntityId: 'comp-source', newEntityId: 'team-new' },
      () =>
        expect(linkComponentToInternalTeamMutate).toHaveBeenCalledWith(
          expect.objectContaining({ componentId: 'comp-source', teamId: 'team-new' }),
        ),
    ],
  ])('%s', async (_name, args, assertion) => {
    await runFlow(args);
    assertion();
  });
});

describe('useCreateRelatedEntity — handleEntityCreated regular mode add-to-view', () => {
  it('adds the new component to the view at the offset right of the source', async () => {
    await runFlow({ entry: triggersEntry, sourceEntityId: 'comp-source', newEntityId: 'comp-new' });
    expect(addComponentToViewMutate).toHaveBeenCalledTimes(1);
    const call = addComponentToViewMutate.mock.calls[0][0];
    expect(call.viewId).toBe('v1');
    expect(call.request.componentId).toBe('comp-new');
    expect(call.request.x).toBeGreaterThan(100);
    expect(call.request.y).toBe(200);
  });

  it('clears pending after a successful create', async () => {
    const result = await runFlow({ entry: triggersEntry, sourceEntityId: 'comp-source', newEntityId: 'comp-new' });
    expect(result.current.pending).toBeNull();
  });

  it('uses addCapabilityToView for capability-parent flows', async () => {
    await runFlow({ entry: capabilityParentEntry, sourceEntityId: 'cap-source', newEntityId: 'cap-new' });
    expect(addCapabilityToViewMutate).toHaveBeenCalled();
  });

  it('uses addOriginEntityToView when creating an acquired-entity from a component', async () => {
    await runFlow({
      entry: acquiredViaFromComponent,
      sourceEntityId: 'comp-source',
      newEntityId: 'acq-new',
    });
    expect(addOriginEntityToViewMutate).toHaveBeenCalled();
  });
});

describe('useCreateRelatedEntity — relation failure rollback', () => {
  it('keeps the just-created entity on the canvas by still adding it to the view (spec rule 13)', async () => {
    createRelationMutate.mockRejectedValueOnce(new Error('relation not allowed'));

    const result = await runFlow({ entry: triggersEntry, sourceEntityId: 'comp-source', newEntityId: 'comp-new' });

    expect(toastError).toHaveBeenCalledWith(expect.stringMatching(/Component \(triggers\)/));
    expect(addComponentToViewMutate).toHaveBeenCalledTimes(1);
    expect(addComponentToViewMutate.mock.calls[0][0].request.componentId).toBe('comp-new');
    expect(result.current.pending).toBeNull();
  });

  it('aborts and does not add to view when the relationType is unknown to the planner', async () => {
    const unknownEntry: RelatedLink = {
      href: '/api/v1/components',
      methods: ['POST'],
      title: 'Mystery (unknown)',
      targetType: 'component',
      relationType: 'totally-unknown-relation-type',
    };

    const result = await runFlow({ entry: unknownEntry, sourceEntityId: 'comp-source', newEntityId: 'comp-new' });

    expect(toastError).toHaveBeenCalledWith(expect.stringMatching(/totally-unknown-relation-type/));
    expect(createRelationMutate).not.toHaveBeenCalled();
    expect(addComponentToViewMutate).not.toHaveBeenCalled();
    expect(result.current.pending).toBeNull();
  });

  it('aborts and does not draft the entity in dynamic mode when the relationType is unknown', async () => {
    enableDynamicMode();
    const unknownEntry: RelatedLink = {
      href: '/api/v1/components',
      methods: ['POST'],
      title: 'Mystery (unknown)',
      targetType: 'component',
      relationType: 'totally-unknown-relation-type',
    };

    await runFlow({ entry: unknownEntry, sourceEntityId: 'comp-source', newEntityId: 'comp-new' });

    expect(toastError).toHaveBeenCalledWith(expect.stringMatching(/totally-unknown-relation-type/));
    const state = useAppStore.getState();
    expect(state.dynamicEntities.map((e) => e.id)).not.toContain('comp-new');
  });
});

describe('useCreateRelatedEntity — dynamic mode', () => {
  it('dispatches the relation to the backend immediately while drafting only the view placement', async () => {
    enableDynamicMode();
    await runFlow({ entry: triggersEntry, sourceEntityId: 'comp-source', newEntityId: 'comp-new' });

    expect(createRelationMutate).toHaveBeenCalledWith(
      expect.objectContaining({
        sourceComponentId: 'comp-source',
        targetComponentId: 'comp-new',
        relationType: 'Triggers',
      }),
    );
    expect(addComponentToViewMutate).not.toHaveBeenCalled();

    const state = useAppStore.getState();
    expect(state.dynamicEntities.map((e) => e.id)).toContain('comp-new');
    expect(state.dynamicPositions['comp-new']).toBeDefined();
    expect(state.dynamicPositions['comp-new'].x).toBeGreaterThan(100);
    expect(state.dynamicPositions['comp-new'].y).toBe(200);
  });

  it('dispatches a Serves relation when the picker entry is component-serves', async () => {
    enableDynamicMode();
    await runFlow({ entry: servesEntry, sourceEntityId: 'comp-source', newEntityId: 'comp-new' });

    expect(createRelationMutate).toHaveBeenCalledWith(
      expect.objectContaining({ relationType: 'Serves' }),
    );
  });

  it('dispatches a capability-parent relation immediately when the source is a capability', async () => {
    enableDynamicMode();
    await runFlow({ entry: capabilityParentEntry, sourceEntityId: 'cap-source', newEntityId: 'cap-new' });

    expect(changeCapabilityParentMutate).toHaveBeenCalledWith(
      expect.objectContaining({ id: 'cap-new', newParentId: 'cap-source' }),
    );
    expect(addCapabilityToViewMutate).not.toHaveBeenCalled();
  });

  it('dispatches origin-acquired-via immediately and adds origin entity to draft when creating from a component', async () => {
    enableDynamicMode();
    await runFlow({
      entry: acquiredViaFromComponent,
      sourceEntityId: 'comp-source',
      newEntityId: 'acq-new',
    });

    expect(linkComponentToAcquiredEntityMutate).toHaveBeenCalledWith(
      expect.objectContaining({ componentId: 'comp-source', entityId: 'acq-new' }),
    );
    expect(addOriginEntityToViewMutate).not.toHaveBeenCalled();

    const state = useAppStore.getState();
    expect(state.dynamicEntities.map((e) => e.id)).toContain('acq-new');
  });
});
