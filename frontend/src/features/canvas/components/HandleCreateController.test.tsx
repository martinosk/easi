import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen } from '@testing-library/react';
import { ReactFlowProvider } from '@xyflow/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { HATEOASLinks, ViewId } from '../../../api/types';
import { useAppStore } from '../../../store/appStore';
import { MantineTestWrapper } from '../../../test/helpers/mantineTestWrapper';
import { HandleCreateController } from './HandleCreateController';

const componentsData: { id: string; name: string; _links: HATEOASLinks }[] = [];

vi.mock('../../components/hooks/useComponents', () => ({
  useComponents: () => ({ data: componentsData }),
  useCreateComponent: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

vi.mock('../../capabilities/hooks/useCapabilities', () => ({
  useCapabilities: () => ({ data: [] }),
  useChangeCapabilityParent: () => ({ mutateAsync: vi.fn() }),
  useLinkSystemToCapability: () => ({ mutateAsync: vi.fn() }),
  useCreateCapability: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useUpdateCapabilityMetadata: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

vi.mock('../../origin-entities/hooks/useAcquiredEntities', () => ({
  useAcquiredEntitiesQuery: () => ({ data: [] }),
  useCreateAcquiredEntity: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

vi.mock('../../origin-entities/hooks/useVendors', () => ({
  useVendorsQuery: () => ({ data: [] }),
  useCreateVendor: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

vi.mock('../../origin-entities/hooks/useInternalTeams', () => ({
  useInternalTeamsQuery: () => ({ data: [] }),
  useCreateInternalTeam: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

vi.mock('../../origin-entities/hooks', () => ({
  useLinkComponentToAcquiredEntity: () => ({ mutateAsync: vi.fn() }),
  useLinkComponentToVendor: () => ({ mutateAsync: vi.fn() }),
  useLinkComponentToInternalTeam: () => ({ mutateAsync: vi.fn() }),
}));

vi.mock('../../relations/hooks/useRelations', () => ({
  useCreateRelation: () => ({ mutateAsync: vi.fn() }),
}));

vi.mock('../../views/hooks/useViews', () => ({
  useAddComponentToView: () => ({ mutateAsync: vi.fn() }),
  useAddCapabilityToView: () => ({ mutateAsync: vi.fn() }),
  useAddOriginEntityToView: () => ({ mutateAsync: vi.fn() }),
}));

vi.mock('../../views/hooks/useCurrentView', () => ({
  useCurrentView: () => ({
    currentView: { id: 'v1', _links: { edit: { href: '/x', method: 'PUT' } } },
    currentViewId: 'v1' as ViewId,
    isLoading: false,
    error: null,
  }),
}));

vi.mock('../../../hooks/useMaturityScale', () => ({
  useMaturityScale: () => ({ data: { sections: [] } }),
}));

vi.mock('../../../hooks/useMetadata', () => ({
  useStatuses: () => ({ data: [], isLoading: false }),
  useMaturityLevels: () => ({ data: [], isLoading: false }),
}));

const linksWithRelated = (entries: unknown): HATEOASLinks =>
  ({ self: { href: '/x', method: 'GET' }, 'x-related': entries }) as unknown as HATEOASLinks;

function setupComponent(id: string, related: unknown[]) {
  componentsData.length = 0;
  componentsData.push({ id, name: 'A', _links: linksWithRelated(related) });
}

function renderController() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <ReactFlowProvider>
        <MantineTestWrapper>
          <div data-testid="canvas-fake">
            <div data-id="comp-1">
              <div className="react-flow__handle" data-handlepos="right" data-testid="handle-right" />
            </div>
          </div>
          <HandleCreateController />
        </MantineTestWrapper>
      </ReactFlowProvider>
    </QueryClientProvider>,
  );
}

beforeEach(() => {
  componentsData.length = 0;
  useAppStore.setState({
    currentViewId: 'v1' as ViewId,
    dynamicViewId: null,
    dynamicEntities: [],
    dynamicOriginal: null,
    dynamicPositions: {},
    draftsByView: {},
  });
});

afterEach(() => vi.clearAllMocks());

describe('HandleCreateController', () => {
  it('opens the picker on a handle click and lists POST-capable entries', async () => {
    setupComponent('comp-1', [
      {
        href: '/api/v1/components',
        methods: ['POST'],
        title: 'Component (related)',
        targetType: 'component',
        relationType: 'component-relation',
      },
    ]);

    renderController();

    const handle = screen.getByTestId('handle-right');
    fireEvent.mouseDown(handle, { clientX: 100, clientY: 50 });
    fireEvent.mouseUp(handle, { clientX: 100, clientY: 50 });

    expect(await screen.findByRole('menuitem', { name: 'Component (related)' })).toBeInTheDocument();
  });

  it('does not open any picker when there are no POST-capable entries', () => {
    setupComponent('comp-1', [
      {
        href: '/api/v1/components',
        methods: ['GET'],
        title: 'Read only',
        targetType: 'component',
        relationType: 'component-relation',
      },
    ]);

    renderController();
    const handle = screen.getByTestId('handle-right');
    fireEvent.mouseDown(handle, { clientX: 100, clientY: 50 });
    fireEvent.mouseUp(handle, { clientX: 100, clientY: 50 });

    expect(screen.queryByRole('menuitem')).toBeNull();
  });
});
