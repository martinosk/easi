import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import toast from 'react-hot-toast';
import { useParams } from 'react-router-dom';
import type { BusinessDomain, BusinessDomainId, Capability, CapabilityId, ComponentId } from '../../../api/types';
import { useMaturityColorScale } from '../../../hooks/useMaturityColorScale';
import { clearParams, deepLinkParams, getParamValue } from '../../../lib/deepLinks';
import { useUserStore } from '../../../store/userStore';
import type { DomainBoardViewModel } from './domainBoardViewModel';
import { flattenViewModelCapabilities } from './domainBoardViewModel';
import { useCapabilityContextMenu } from './useCapabilityContextMenu';
import { useCapabilitySelection } from './useCapabilitySelection';
import { useDomainBoardData } from './useDomainBoardData';
import { useAssociateCapability, useDissociateCapability } from './useDomainCapabilities';
import { useDomainContextMenu } from './useDomainContextMenu';
import { useDomainDialogManager } from './useDomainDialogManager';
import { useDragHandlers } from './useDragHandlers';
import { useKeyboardShortcuts } from './useKeyboardShortcuts';

const HIGHLIGHT_DURATION_MS = 2000;

function findL1Ancestor(capability: Capability, allCapabilities: Capability[]): Capability {
  const byId = new Map(allCapabilities.map((c) => [c.id, c]));
  let current = capability;
  while (current.parentId) {
    const parent = byId.get(current.parentId);
    if (!parent) break;
    current = parent;
  }
  return current;
}

function useDomainDeepLink(domains: BusinessDomain[], isLoading: boolean, onFound: (domain: BusinessDomain) => void) {
  const { domainId: pathDomainId } = useParams<{ domainId?: string }>();
  const processedRef = useRef(false);

  useEffect(() => {
    if (isLoading || processedRef.current) return;

    const domainIdFromUrl = pathDomainId ?? getParamValue(deepLinkParams.DOMAIN.param);
    if (!domainIdFromUrl) return;

    processedRef.current = true;
    const linkedDomain = domains.find((d) => d.id === domainIdFromUrl);

    if (linkedDomain) {
      onFound(linkedDomain);
    } else {
      toast.error('The linked domain does not exist');
    }

    clearParams([deepLinkParams.DOMAIN.param]);
  }, [domains, isLoading, onFound, pathDomainId]);
}

function useCapabilityDeepLink(
  allCapabilities: Capability[],
  isLoading: boolean,
  onFound: (capability: Capability) => void,
) {
  const processedRef = useRef(false);

  useEffect(() => {
    if (isLoading || processedRef.current) return;
    if (allCapabilities.length === 0) return;

    const capabilityIdFromUrl = getParamValue(deepLinkParams.CAPABILITY.param);
    if (!capabilityIdFromUrl) return;

    processedRef.current = true;
    const linkedCapability = allCapabilities.find((c) => c.id === capabilityIdFromUrl);

    if (linkedCapability) {
      onFound(linkedCapability);
    } else {
      toast.error('The linked capability does not exist');
    }

    clearParams([deepLinkParams.CAPABILITY.param]);
  }, [allCapabilities, isLoading, onFound]);
}

function useCapabilityDrawerState() {
  const [selectedCapability, setSelectedCapability] = useState<Capability | null>(null);
  const [selectedCapabilityDomainId, setSelectedCapabilityDomainId] = useState<BusinessDomainId | null>(null);
  const [selectedComponentId, setSelectedComponentId] = useState<ComponentId | null>(null);

  const clearCapabilityDetails = useCallback(() => {
    setSelectedCapability(null);
    setSelectedCapabilityDomainId(null);
  }, []);

  const clearSelectedComponent = useCallback(() => {
    setSelectedComponentId(null);
  }, []);

  const handleApplicationClick = useCallback((componentId: ComponentId) => {
    setSelectedComponentId(componentId);
    setSelectedCapability(null);
  }, []);

  const openCapabilityDrawer = useCallback((domainId: BusinessDomainId, capability: Capability) => {
    setSelectedCapability(capability);
    setSelectedCapabilityDomainId(domainId);
    setSelectedComponentId(null);
  }, []);

  return {
    selectedCapability,
    selectedCapabilityDomainId,
    selectedComponentId,
    clearCapabilityDetails,
    clearSelectedComponent,
    handleApplicationClick,
    openCapabilityDrawer,
  };
}

function useActiveDomainSelection(
  boardDomainsById: Map<BusinessDomainId, DomainBoardViewModel>,
  openCapabilityDrawer: (domainId: BusinessDomainId, capability: Capability) => void,
) {
  const [activeDomainId, setActiveDomainId] = useState<BusinessDomainId | null>(null);

  const activeDomainCapabilities = useMemo(() => {
    const viewModel = activeDomainId ? boardDomainsById.get(activeDomainId) : undefined;
    return viewModel ? flattenViewModelCapabilities(viewModel) : [];
  }, [activeDomainId, boardDomainsById]);

  const activeDomainAssignedCapabilities = useMemo(() => {
    const viewModel = activeDomainId ? boardDomainsById.get(activeDomainId) : undefined;
    return viewModel ? viewModel.assignedCapabilities : [];
  }, [activeDomainId, boardDomainsById]);

  const activeDomainIdRef = useRef<BusinessDomainId | null>(null);

  const selection = useCapabilitySelection(activeDomainCapabilities, (capability) => {
    if (activeDomainIdRef.current) openCapabilityDrawer(activeDomainIdRef.current, capability);
  });

  const switchActiveDomain = useCallback(
    (domainId: BusinessDomainId) => {
      if (activeDomainIdRef.current !== domainId) {
        selection.clearSelection();
        setActiveDomainId(domainId);
      }
      activeDomainIdRef.current = domainId;
    },
    [selection],
  );

  const handleCapabilityClick = useCallback(
    (domainId: BusinessDomainId, capability: Capability, event: React.MouseEvent) => {
      switchActiveDomain(domainId);
      selection.handleCapabilityClick(capability, event);
    },
    [switchActiveDomain, selection],
  );

  return {
    activeDomainId,
    setActiveDomainId,
    activeDomainCapabilities,
    activeDomainAssignedCapabilities,
    selection,
    switchActiveDomain,
    handleCapabilityClick,
  };
}

interface CapabilityBoardInteractionsParams {
  activeDomainId: BusinessDomainId | null;
  activeDomainCapabilities: Capability[];
  activeDomainAssignedCapabilities: Capability[];
  boardDomainsById: Map<BusinessDomainId, DomainBoardViewModel>;
  selection: ReturnType<typeof useCapabilitySelection>;
  switchActiveDomain: (domainId: BusinessDomainId) => void;
  refetchDomain: (domainId: BusinessDomainId) => Promise<void>;
}

function useCapabilityBoardInteractions({
  activeDomainId,
  activeDomainCapabilities,
  activeDomainAssignedCapabilities,
  boardDomainsById,
  selection,
  switchActiveDomain,
  refetchDomain,
}: CapabilityBoardInteractionsParams) {
  const associateMutation = useAssociateCapability();
  const dissociateMutation = useDissociateCapability();

  const capabilityContextMenuBase = useCapabilityContextMenu({
    capabilities: activeDomainCapabilities,
    domainCapabilities: activeDomainAssignedCapabilities,
    dissociateCapability: async (capability) => {
      if (!activeDomainId) return;
      await dissociateMutation.mutateAsync({ domainId: activeDomainId, capability });
    },
    refetch: async () => {
      if (activeDomainId) await refetchDomain(activeDomainId);
    },
    selectedCapabilities: selection.selectedCapabilities,
    setSelectedCapabilities: selection.setSelectedCapabilities,
  });

  const handleCapabilityContextMenu = useCallback(
    (domainId: BusinessDomainId, capability: Capability, event: React.MouseEvent) => {
      switchActiveDomain(domainId);
      capabilityContextMenuBase.handleCapabilityContextMenu(capability, event);
    },
    [switchActiveDomain, capabilityContextMenuBase],
  );

  const isCapabilityAssignedToDomain = useCallback(
    (domainId: BusinessDomainId, capabilityId: CapabilityId) => {
      const viewModel = boardDomainsById.get(domainId);
      return viewModel ? viewModel.assignedCapabilities.some((c) => c.id === capabilityId) : false;
    },
    [boardDomainsById],
  );

  const dragHandlers = useDragHandlers({
    associateCapability: async (domainId, capabilityId) => {
      await associateMutation.mutateAsync({ domainId, capabilityId });
    },
    isCapabilityAssignedToDomain,
    refetchDomain,
  });

  useKeyboardShortcuts({
    hasSelection: selection.selectedCapabilities.size > 0,
    onSelectAll: selection.selectAllL1Capabilities,
    onClearSelection: selection.clearSelection,
  });

  return {
    capabilityContextMenu: { ...capabilityContextMenuBase, handleCapabilityContextMenu },
    dragHandlers,
  };
}

interface BusinessDomainsDeepLinksParams {
  domains: BusinessDomain[];
  isLoading: boolean;
  allCapabilities: Capability[];
  boardDomains: DomainBoardViewModel[];
  switchActiveDomain: (domainId: BusinessDomainId) => void;
  openCapabilityDrawer: (domainId: BusinessDomainId, capability: Capability) => void;
}

function useBusinessDomainsDeepLinks({
  domains,
  isLoading,
  allCapabilities,
  boardDomains,
  switchActiveDomain,
  openCapabilityDrawer,
}: BusinessDomainsDeepLinksParams) {
  const [highlightedDomainId, setHighlightedDomainId] = useState<BusinessDomainId | null>(null);
  const [forceOpenL1Ids, setForceOpenL1Ids] = useState<Set<CapabilityId>>(new Set());

  useEffect(() => {
    if (!highlightedDomainId) return;
    const timer = setTimeout(() => setHighlightedDomainId(null), HIGHLIGHT_DURATION_MS);
    return () => clearTimeout(timer);
  }, [highlightedDomainId]);

  useDomainDeepLink(domains, isLoading, (domain) => {
    setHighlightedDomainId(domain.id);
  });

  useCapabilityDeepLink(allCapabilities, isLoading, (capability) => {
    const l1Ancestor = findL1Ancestor(capability, allCapabilities);
    const owningDomain = boardDomains.find((vm) => vm.assignedCapabilities.some((c) => c.id === l1Ancestor.id));

    if (!owningDomain) {
      toast.error('The linked capability is not assigned to a business domain');
      return;
    }

    switchActiveDomain(owningDomain.domain.id);
    setForceOpenL1Ids(new Set([l1Ancestor.id]));
    setHighlightedDomainId(owningDomain.domain.id);
    openCapabilityDrawer(owningDomain.domain.id, capability);
  });

  return { highlightedDomainId, forceOpenL1Ids };
}

function useBoardFilters() {
  const userRole = useUserStore((state) => state.user?.role);
  const { getColorForValue } = useMaturityColorScale();
  const [searchQuery, setSearchQuery] = useState('');
  const [assignRailOpen, setAssignRailOpen] = useState(false);

  return {
    showAssignRail: userRole !== 'stakeholder',
    getColorForValue,
    searchQuery,
    setSearchQuery,
    assignRailOpen,
    toggleAssignRail: () => setAssignRailOpen((open) => !open),
  };
}

function useBoardIndexes(boardDomains: DomainBoardViewModel[]) {
  const boardDomainsById = useMemo(() => new Map(boardDomains.map((vm) => [vm.domain.id, vm])), [boardDomains]);

  const globalAssignedCapabilityIds = useMemo(
    () => new Set(boardDomains.flatMap((vm) => vm.assignedCapabilities.map((c) => c.id))),
    [boardDomains],
  );

  return { boardDomainsById, globalAssignedCapabilityIds };
}

interface DomainDialogAndMenuParams {
  createDomain: (name: string, description?: string, domainArchitectId?: string) => Promise<BusinessDomain>;
  updateDomain: (
    domain: BusinessDomain,
    name: string,
    description?: string,
    domainArchitectId?: string,
  ) => Promise<BusinessDomain>;
  deleteDomain: (domain: BusinessDomain) => Promise<void>;
  activeDomainId: BusinessDomainId | null;
  setActiveDomainId: (domainId: BusinessDomainId | null) => void;
  selectedCapabilityDomainId: BusinessDomainId | null;
  clearCapabilityDetails: () => void;
}

function useDomainDialogAndMenu({
  createDomain,
  updateDomain,
  deleteDomain,
  activeDomainId,
  setActiveDomainId,
  selectedCapabilityDomainId,
  clearCapabilityDetails,
}: DomainDialogAndMenuParams) {
  const dialogManager = useDomainDialogManager({
    createDomain,
    updateDomain,
    deleteDomain,
    onDomainDeleted: (deletedId) => {
      if (activeDomainId === deletedId) setActiveDomainId(null);
      if (selectedCapabilityDomainId === deletedId) clearCapabilityDetails();
    },
  });

  const domainContextMenu = useDomainContextMenu({
    onEdit: dialogManager.handleEditClick,
    onDelete: dialogManager.handleDeleteClick,
  });

  return { dialogManager, domainContextMenu };
}

function useSelectedCapabilityDetails(
  selectedCapability: Capability | null,
  selectedCapabilityDomainId: BusinessDomainId | null,
  boardDomainsById: Map<BusinessDomainId, DomainBoardViewModel>,
  allCapabilities: Capability[],
) {
  const selectedDomainViewModel = selectedCapabilityDomainId
    ? boardDomainsById.get(selectedCapabilityDomainId)
    : undefined;

  return {
    selectedDomain: selectedDomainViewModel?.domain ?? null,
    selectedL1Name: selectedCapability ? findL1Ancestor(selectedCapability, allCapabilities).name : null,
    getRealizationsForSelectedCapability: selectedDomainViewModel?.getRealizationsForCapability ?? (() => []),
  };
}

interface DomainBoardInteractionsParams {
  boardDomains: DomainBoardViewModel[];
  domains: BusinessDomain[];
  isLoading: boolean;
  allCapabilities: Capability[];
  refetchDomain: (domainId: BusinessDomainId) => Promise<void>;
  drawer: ReturnType<typeof useCapabilityDrawerState>;
}

function useDomainBoardInteractions({
  boardDomains,
  domains,
  isLoading,
  allCapabilities,
  refetchDomain,
  drawer,
}: DomainBoardInteractionsParams) {
  const { boardDomainsById, globalAssignedCapabilityIds } = useBoardIndexes(boardDomains);
  const activeDomain = useActiveDomainSelection(boardDomainsById, drawer.openCapabilityDrawer);

  const { capabilityContextMenu, dragHandlers } = useCapabilityBoardInteractions({
    activeDomainId: activeDomain.activeDomainId,
    activeDomainCapabilities: activeDomain.activeDomainCapabilities,
    activeDomainAssignedCapabilities: activeDomain.activeDomainAssignedCapabilities,
    boardDomainsById,
    selection: activeDomain.selection,
    switchActiveDomain: activeDomain.switchActiveDomain,
    refetchDomain,
  });

  const { highlightedDomainId, forceOpenL1Ids } = useBusinessDomainsDeepLinks({
    domains,
    isLoading,
    allCapabilities,
    boardDomains,
    switchActiveDomain: activeDomain.switchActiveDomain,
    openCapabilityDrawer: drawer.openCapabilityDrawer,
  });

  const { selectedDomain, selectedL1Name, getRealizationsForSelectedCapability } = useSelectedCapabilityDetails(
    drawer.selectedCapability,
    drawer.selectedCapabilityDomainId,
    boardDomainsById,
    allCapabilities,
  );

  return {
    activeDomain,
    globalAssignedCapabilityIds,
    capabilityContextMenu,
    dragHandlers,
    highlightedDomainId,
    forceOpenL1Ids,
    selectedDomain,
    selectedL1Name,
    getRealizationsForSelectedCapability,
  };
}

export function useBusinessDomainsPage() {
  const board = useDomainBoardData();
  const {
    domains,
    boardDomains,
    canCreateDomain,
    isLoading,
    error,
    allCapabilities,
    refetchDomain,
    createDomain,
    updateDomain,
    deleteDomain,
  } = board;

  const filters = useBoardFilters();
  const drawer = useCapabilityDrawerState();

  const {
    activeDomain,
    globalAssignedCapabilityIds,
    capabilityContextMenu,
    dragHandlers,
    highlightedDomainId,
    forceOpenL1Ids,
    selectedDomain,
    selectedL1Name,
    getRealizationsForSelectedCapability,
  } = useDomainBoardInteractions({ boardDomains, domains, isLoading, allCapabilities, refetchDomain, drawer });

  const { dialogManager, domainContextMenu } = useDomainDialogAndMenu({
    createDomain,
    updateDomain,
    deleteDomain,
    activeDomainId: activeDomain.activeDomainId,
    setActiveDomainId: activeDomain.setActiveDomainId,
    selectedCapabilityDomainId: drawer.selectedCapabilityDomainId,
    clearCapabilityDetails: drawer.clearCapabilityDetails,
  });

  return {
    boardDomains,
    canCreateDomain,
    isLoading,
    error,
    allCapabilities,
    globalAssignedCapabilityIds,
    ...filters,
    selectedCapability: drawer.selectedCapability,
    selectedDomain,
    selectedL1Name,
    getRealizationsForSelectedCapability,
    selectedComponentId: drawer.selectedComponentId,
    selectedCapabilities: activeDomain.selection.selectedCapabilities,
    highlightedDomainId,
    forceOpenL1Ids,
    dialogManager,
    domainContextMenu,
    capabilityContextMenu,
    dragHandlers,
    handleCapabilityClick: activeDomain.handleCapabilityClick,
    clearCapabilityDetails: drawer.clearCapabilityDetails,
    handleApplicationClick: drawer.handleApplicationClick,
    clearSelectedComponent: drawer.clearSelectedComponent,
  };
}
