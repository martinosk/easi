import { vi } from 'vitest';
import type { BusinessDomainId, Capability, CapabilityId, CapabilityRealizationsGroup } from '../../../api/types';
import { buildBusinessDomain } from '../../../test/helpers';
import { buildCapabilityAt as cap } from '../../../test/helpers/entityBuilders';
import { buildCapabilityTree } from '../../capabilities/hooks/useCapabilityTree';
import { buildDomainBoardViewModel } from '../hooks/domainBoardViewModel';
import type { useBusinessDomainsPage } from '../hooks/useBusinessDomainsPage';
import { buildJourneyIndex } from '../lens/journeyIndex';

export type BusinessDomainsHookData = ReturnType<typeof useBusinessDomainsPage>;

export function buildViewModel(
  name: string,
  id: string,
  capabilities?: Capability[],
  realizationGroups: CapabilityRealizationsGroup[] = [],
) {
  const domain = buildBusinessDomain({ id: id as BusinessDomainId, name });
  const assigned = capabilities ?? [cap(`${id}-l1`, `${name} Capability`, 'L1')];
  return buildDomainBoardViewModel({
    domain,
    assignedCapabilities: assigned,
    tree: buildCapabilityTree(assigned),
    realizationGroups,
    isLoading: false,
  });
}

export function buildHookData(overrides: Partial<BusinessDomainsHookData> = {}): BusinessDomainsHookData {
  return {
    boardDomains: [buildViewModel('Ferry Freight', 'domain-1'), buildViewModel('Logistics', 'domain-2')],
    journeyIndex: buildJourneyIndex({ journeys: [], capabilityDomainNames: new Map() }),
    lens: 'now',
    setLens: vi.fn(),
    changesOnly: false,
    setChangesOnly: vi.fn(),
    openCapabilityById: vi.fn(),
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
    dialogManager: { handleCreateClick: vi.fn() } as unknown as BusinessDomainsHookData['dialogManager'],
    domainContextMenu: { handleContextMenu: vi.fn() } as unknown as BusinessDomainsHookData['domainContextMenu'],
    capabilityContextMenu: {
      handleCapabilityContextMenu: vi.fn(),
    } as unknown as BusinessDomainsHookData['capabilityContextMenu'],
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
  } as BusinessDomainsHookData;
}
