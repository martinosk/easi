import { useState, useCallback, useEffect, useRef } from 'react';
import toast from 'react-hot-toast';
import { useBusinessDomains } from './useBusinessDomains';
import { useDomainCapabilities } from './useDomainCapabilities';
import { useCapabilityTree } from './useCapabilityTree';
import { useGridPositions } from './useGridPositions';
import { usePersistedDepth } from './usePersistedDepth';
import { useSidebarState } from './useSidebarState';
import { useDomainDialogManager } from './useDomainDialogManager';
import { useDragHandlers } from './useDragHandlers';
import { useCapabilityFiltering } from './useCapabilityFiltering';
import { useDomainContextMenu } from './useDomainContextMenu';
import { useApplicationSettings } from './useApplicationSettings';
import { useCapabilityRealizations } from './useCapabilityRealizations';
import { useCapabilitySelection } from './useCapabilitySelection';
import { useCapabilityContextMenu } from './useCapabilityContextMenu';
import { useKeyboardShortcuts } from './useKeyboardShortcuts';
import { getParamValue, clearParams, deepLinkParams } from '../../../lib/deepLinks';
import type { BusinessDomain, Capability, ComponentId } from '../../../api/types';

export function useBusinessDomainsPage() {
  const [visualizedDomain, setVisualizedDomain] = useState<BusinessDomain | null>(null);
  const [selectedCapability, setSelectedCapability] = useState<Capability | null>(null);
  const [selectedComponentId, setSelectedComponentId] = useState<ComponentId | null>(null);
  const [depth, setDepth] = usePersistedDepth();
  const {
    showApplications,
    setShowApplications,
  } = useApplicationSettings();

  const { domains, isLoading, error, createDomain, updateDomain, deleteDomain } = useBusinessDomains();
  const { tree, isLoading: treeLoading } = useCapabilityTree();
  const deepLinkProcessedRef = useRef(false);

  useEffect(() => {
    if (isLoading || deepLinkProcessedRef.current) return;

    const domainIdFromUrl = getParamValue(deepLinkParams.DOMAIN.param);
    if (!domainIdFromUrl) return;

    deepLinkProcessedRef.current = true;
    const linkedDomain = domains.find(d => d.id === domainIdFromUrl);

    if (linkedDomain) {
      setVisualizedDomain(linkedDomain);
    } else {
      toast.error('The linked domain does not exist');
    }

    clearParams([deepLinkParams.DOMAIN.param]);
  }, [domains, isLoading]);

  const sidebarState = useSidebarState();

  const dialogManager = useDomainDialogManager({
    createDomain,
    updateDomain,
    deleteDomain,
    onDomainDeleted: (deletedId) => {
      if (visualizedDomain?.id === deletedId) {
        setVisualizedDomain(null);
      }
    },
  });

  const { positions, updatePosition } = useGridPositions(visualizedDomain?.id ?? null);

  const {
    capabilities,
    isLoading: capabilitiesLoading,
    associateCapability,
    dissociateCapability,
    refetch: refetchCapabilities,
  } = useDomainCapabilities(visualizedDomain?.id);

  const filtering = useCapabilityFiltering(tree, capabilities);

  const { getRealizationsForCapability, refetch: refetchRealizations } = useCapabilityRealizations(
    showApplications,
    visualizedDomain?.id ?? null,
    depth
  );

  const handleApplicationClick = useCallback((componentId: ComponentId) => {
    setSelectedComponentId(componentId);
    setSelectedCapability(null);
  }, []);

  const clearSelectedComponent = useCallback(() => {
    setSelectedComponentId(null);
  }, []);

  const dragHandlers = useDragHandlers({
    domainId: visualizedDomain?.id ?? null,
    capabilities,
    assignedCapabilityIds: filtering.assignedCapabilityIds,
    positions,
    updatePosition,
    associateCapability,
    refetchCapabilities,
    refetchRealizations,
  });

  const domainContextMenu = useDomainContextMenu({
    onEdit: dialogManager.handleEditClick,
    onDelete: dialogManager.handleDeleteClick,
  });

  const handleVisualizeClick = (domain: BusinessDomain) => {
    setVisualizedDomain(domain);
    setSelectedCapability(null);
  };

  const onRegularCapabilityClick = useCallback((capability: Capability) => {
    setSelectedCapability(capability);
    setSelectedComponentId(null);
  }, []);

  const clearCapabilityDetails = useCallback(() => {
    setSelectedCapability(null);
  }, []);

  const {
    selectedCapabilities,
    handleCapabilityClick,
    selectAllL1Capabilities,
    clearSelection,
    setSelectedCapabilities,
  } = useCapabilitySelection(filtering.capabilitiesWithDescendants, onRegularCapabilityClick);

  const capabilityContextMenu = useCapabilityContextMenu({
    capabilities: filtering.capabilitiesWithDescendants,
    domainCapabilities: capabilities,
    dissociateCapability,
    refetch: refetchCapabilities,
    selectedCapabilities,
    setSelectedCapabilities,
  });

  useKeyboardShortcuts({
    hasSelection: selectedCapabilities.size > 0,
    onSelectAll: selectAllL1Capabilities,
    onClearSelection: clearSelection,
  });

  return {
    domains,
    isLoading,
    error,
    tree,
    treeLoading,
    visualizedDomain,
    selectedCapability,
    selectedComponentId,
    depth,
    setDepth,
    showApplications,
    setShowApplications,
    sidebarState,
    dialogManager,
    positions,
    capabilities,
    capabilitiesLoading,
    filtering,
    dragHandlers,
    domainContextMenu,
    capabilityContextMenu,
    selectedCapabilities,
    handleVisualizeClick,
    handleCapabilityClick,
    clearCapabilityDetails,
    getRealizationsForCapability,
    handleApplicationClick,
    clearSelectedComponent,
  };
}
