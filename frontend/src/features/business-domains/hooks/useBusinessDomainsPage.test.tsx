import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import React from 'react';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { toBusinessDomainId, toComponentId } from '../../../api/types';
import { buildBusinessDomain } from '../../../test/helpers';
import { buildCapabilityAt as cap } from '../../../test/helpers/entityBuilders';
import { buildCapabilityTree } from '../../capabilities/hooks/useCapabilityTree';
import { buildJourneyIndex } from '../lens/journeyIndex';
import { buildDomainBoardViewModel } from './domainBoardViewModel';
import { useBusinessDomainsPage } from './useBusinessDomainsPage';
import { useDomainBoardData } from './useDomainBoardData';

vi.mock('./useDomainBoardData', () => ({
  useDomainBoardData: vi.fn(),
}));

vi.mock('../../../hooks/useMaturityColorScale', () => ({
  useMaturityColorScale: () => ({
    getColorForValue: () => '#000000',
    getSectionNameForValue: () => 'Genesis',
    getBaseSectionColor: () => '#000000',
    bounds: { min: 0, max: 100 },
  }),
}));

let mockRole = 'domain_architect';
vi.mock('../../../store/userStore', () => ({
  useUserStore: (selector: (state: { user: { role: string } }) => unknown) => selector({ user: { role: mockRole } }),
}));

const domain = buildBusinessDomain({ id: toBusinessDomainId('domain-1'), name: 'Ferry Freight' });
const l1 = cap('l1-a', 'Ferry Booking', 'L1');
const l2 = cap('l2-a1', 'Booking Management', 'L2', 'l1-a');

function buildBoardData() {
  const tree = buildCapabilityTree([l1, l2]);
  const viewModel = buildDomainBoardViewModel({
    domain,
    assignedCapabilities: [l1],
    tree,
    realizationGroups: [],
    isLoading: false,
  });
  return {
    domains: [domain],
    boardDomains: [viewModel],
    journeyIndex: buildJourneyIndex({ journeys: [], capabilityDomainNames: new Map() }),
    canCreateDomain: true,
    isLoading: false,
    error: null,
    tree,
    treeLoading: false,
    allCapabilities: [l1, l2],
    refetchDomain: vi.fn().mockResolvedValue(undefined),
    createDomain: vi.fn(),
    updateDomain: vi.fn(),
    deleteDomain: vi.fn(),
  };
}

function renderPage() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={[`${window.location.pathname}${window.location.search}`]}>{children}</MemoryRouter>
    </QueryClientProvider>
  );
  return renderHook(() => useBusinessDomainsPage(), { wrapper });
}

const clickEvent = () =>
  ({
    shiftKey: false,
    ctrlKey: false,
    metaKey: false,
    preventDefault: vi.fn(),
    stopPropagation: vi.fn(),
  }) as unknown as React.MouseEvent;

describe('useBusinessDomainsPage', () => {
  beforeEach(() => {
    mockRole = 'domain_architect';
    window.history.replaceState({}, '', '/business-domains');
    vi.mocked(useDomainBoardData).mockReturnValue(buildBoardData());
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('opens the capability drawer for the right domain and L1 breadcrumb on card click', () => {
    const { result } = renderPage();

    act(() => {
      result.current.handleCapabilityClick(domain.id, l2, clickEvent());
    });

    expect(result.current.selectedCapability).toEqual(l2);
    expect(result.current.selectedDomain).toEqual(domain);
    expect(result.current.selectedL1Name).toBe('Ferry Booking');
  });

  it('opens the application drawer on chip click and closes the capability drawer', () => {
    const { result } = renderPage();

    act(() => {
      result.current.handleCapabilityClick(domain.id, l2, clickEvent());
    });
    expect(result.current.selectedCapability).not.toBeNull();

    act(() => {
      result.current.handleApplicationClick(toComponentId('comp-1'));
    });

    expect(result.current.selectedComponentId).toBe('comp-1');
    expect(result.current.selectedCapability).toBeNull();
  });

  it('clearCapabilityDetails closes the drawer', () => {
    const { result } = renderPage();

    act(() => {
      result.current.handleCapabilityClick(domain.id, l2, clickEvent());
    });
    act(() => {
      result.current.clearCapabilityDetails();
    });

    expect(result.current.selectedCapability).toBeNull();
    expect(result.current.selectedDomain).toBeNull();
  });

  it('hides the assign rail for the stakeholder role', () => {
    mockRole = 'stakeholder';
    const { result } = renderPage();

    expect(result.current.showAssignRail).toBe(false);
  });

  it('shows the assign rail for non-stakeholder roles', () => {
    mockRole = 'domain_architect';
    const { result } = renderPage();

    expect(result.current.showAssignRail).toBe(true);
  });

  it('resolves a ?capability= deep link to its owning domain, expands the L1 ancestor, and opens the drawer', async () => {
    window.history.replaceState({}, '', '/business-domains?capability=l2-a1');
    const { result } = renderPage();

    await waitFor(() => expect(result.current.selectedCapability).toEqual(l2));

    expect(result.current.selectedDomain).toEqual(domain);
    expect(result.current.forceOpenL1Ids.has(l1.id)).toBe(true);
    expect(result.current.highlightedDomainId).toBe(domain.id);
  });

  it('highlights a ?domain= deep-linked domain and clears the highlight after 2 seconds', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    window.history.replaceState({}, '', '/business-domains?domain=domain-1');
    const { result } = renderPage();

    await waitFor(() => expect(result.current.highlightedDomainId).toBe(domain.id));

    act(() => {
      vi.advanceTimersByTime(2000);
    });

    await waitFor(() => expect(result.current.highlightedDomainId).toBeNull());
  });
});
