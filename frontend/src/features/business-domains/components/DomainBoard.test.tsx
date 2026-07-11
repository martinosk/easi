import { screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { BusinessDomainId, CapabilityId } from '../../../api/types';
import { buildBusinessDomain, renderWithProviders } from '../../../test/helpers';
import { buildCapabilityAt as cap } from '../../../test/helpers/entityBuilders';
import { buildCapabilityTree } from '../../capabilities/hooks/useCapabilityTree';
import { buildDomainBoardViewModel } from '../hooks/domainBoardViewModel';
import type { useBusinessDomainsPage } from '../hooks/useBusinessDomainsPage';
import { DomainBoard } from './DomainBoard';

type HookData = ReturnType<typeof useBusinessDomainsPage>;

function buildViewModel(name: string, id: string) {
  const domain = buildBusinessDomain({ id: id as BusinessDomainId, name });
  const tree = buildCapabilityTree([cap(`${id}-l1`, `${name} Capability`, 'L1')]);
  return buildDomainBoardViewModel({
    domain,
    assignedCapabilities: [cap(`${id}-l1`, `${name} Capability`, 'L1')],
    tree,
    realizationGroups: [],
    isLoading: false,
  });
}

function buildHookData(overrides: Partial<HookData> = {}): HookData {
  return {
    boardDomains: [buildViewModel('Ferry Freight', 'domain-1'), buildViewModel('Logistics', 'domain-2')],
    canCreateDomain: true,
    isLoading: false,
    error: null,
    getColorForValue: () => '#000',
    searchQuery: '',
    setSearchQuery: vi.fn(),
    assignRailOpen: false,
    toggleAssignRail: vi.fn(),
    showAssignRail: true,
    allCapabilities: [],
    globalAssignedCapabilityIds: new Set<CapabilityId>(),
    selectedCapability: null,
    selectedDomain: null,
    selectedL1Name: null,
    getRealizationsForSelectedCapability: () => [],
    selectedComponentId: null,
    selectedCapabilities: new Set<CapabilityId>(),
    highlightedDomainId: null,
    forceOpenL1Ids: new Set<CapabilityId>(),
    dialogManager: { handleCreateClick: vi.fn() } as unknown as HookData['dialogManager'],
    domainContextMenu: { handleContextMenu: vi.fn() } as unknown as HookData['domainContextMenu'],
    capabilityContextMenu: { handleCapabilityContextMenu: vi.fn() } as unknown as HookData['capabilityContextMenu'],
    dragHandlers: {
      activeCapability: null,
      dragOverDomainId: null,
      handleDragStart: vi.fn(),
      handleDragEnd: vi.fn(),
      handleDragOver: vi.fn(),
      handleDragLeave: vi.fn(),
      handleDrop: vi.fn(),
    },
    handleCapabilityClick: vi.fn(),
    clearCapabilityDetails: vi.fn(),
    handleApplicationClick: vi.fn(),
    clearSelectedComponent: vi.fn(),
    ...overrides,
  } as HookData;
}

describe('DomainBoard', () => {
  it('renders the board container and one card per business domain', () => {
    renderWithProviders(<DomainBoard hookData={buildHookData()} />);

    expect(screen.getByTestId('domain-board')).toBeInTheDocument();
    expect(screen.getByTestId('domain-card-domain-1')).toBeInTheDocument();
    expect(screen.getByTestId('domain-card-domain-2')).toBeInTheDocument();
  });

  it('renders the toolbar with search and legend', () => {
    renderWithProviders(<DomainBoard hookData={buildHookData()} />);

    expect(screen.getByTestId('board-search-input')).toBeInTheDocument();
    expect(screen.getByTestId('board-legend')).toBeInTheDocument();
  });

  it('shows an empty-state message when there are no business domains', () => {
    renderWithProviders(<DomainBoard hookData={buildHookData({ boardDomains: [] })} />);

    expect(screen.queryByTestId('domain-board')).not.toBeInTheDocument();
    expect(screen.getByText(/No business domains yet/)).toBeInTheDocument();
  });

  it('hides the assign rail when the toggle is closed', () => {
    renderWithProviders(<DomainBoard hookData={buildHookData({ assignRailOpen: false })} />);
    expect(screen.queryByTestId('assign-rail')).not.toBeInTheDocument();
  });

  it('shows the assign rail when open and the role permits it', () => {
    renderWithProviders(<DomainBoard hookData={buildHookData({ assignRailOpen: true, showAssignRail: true })} />);
    expect(screen.getByTestId('assign-rail')).toBeInTheDocument();
  });

  it('never shows the assign rail for a stakeholder, even if assignRailOpen is somehow true', () => {
    renderWithProviders(<DomainBoard hookData={buildHookData({ assignRailOpen: true, showAssignRail: false })} />);
    expect(screen.queryByTestId('assign-rail')).not.toBeInTheDocument();
  });
});
