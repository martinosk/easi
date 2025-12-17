import { useState, useMemo, useCallback } from 'react';
import { useBusinessDomains } from './useBusinessDomains';
import { useDomainCapabilities } from './useDomainCapabilities';
import { useCapabilityTree } from './useCapabilityTree';
import { useGridPositions } from './useGridPositions';
import { useDragHandlers } from './useDragHandlers';
import { useCapabilityContextMenu } from './useCapabilityContextMenu';
import { useCapabilitySelection } from './useCapabilitySelection';
import { useKeyboardShortcuts } from './useKeyboardShortcuts';
import { usePersistedDepth } from './usePersistedDepth';
import type { BusinessDomainId, Capability } from '../../../api/types';

export function useDomainVisualization(initialDomainId?: BusinessDomainId) {
  const [selectedDomainId, setSelectedDomainId] = useState<BusinessDomainId | null>(initialDomainId ?? null);
  const [selectedCapability, setSelectedCapability] = useState<Capability | null>(null);
  const [depth, setDepth] = usePersistedDepth();
  const { domains, isLoading: domainsLoading } = useBusinessDomains();
  const { tree, isLoading: treeLoading } = useCapabilityTree();
  const { positions, updatePosition } = useGridPositions(selectedDomainId);

  const allCapabilities = useMemo(() => {
    const flatten = (nodes: typeof tree): Capability[] => {
      return nodes.flatMap((node) => [node.capability, ...flatten(node.children)]);
    };
    return flatten(tree);
  }, [tree]);

  const selectedDomain = useMemo(
    () => domains.find((d) => d.id === selectedDomainId),
    [domains, selectedDomainId]
  );

  const {
    capabilities,
    isLoading: capabilitiesLoading,
    associateCapability,
    dissociateCapability,
    refetch: refetchCapabilities,
  } = useDomainCapabilities(selectedDomain?._links.capabilities);

  const assignedCapabilityIds = useMemo(
    () => new Set(capabilities.map((c) => c.id)),
    [capabilities]
  );

  const dragHandlers = useDragHandlers({
    domainId: selectedDomainId,
    capabilities,
    assignedCapabilityIds,
    positions,
    updatePosition,
    associateCapability,
    refetchCapabilities,
  });

  const onRegularClick = useCallback((capability: Capability) => {
    setSelectedCapability(capability);
  }, []);

  const {
    selectedCapabilities,
    handleCapabilityClick,
    selectAllL1Capabilities,
    clearSelection,
    setSelectedCapabilities,
  } = useCapabilitySelection(capabilities, onRegularClick);

  const {
    contextMenu,
    capabilityToDelete,
    capabilitiesToDelete,
    handleCapabilityContextMenu,
    closeContextMenu,
    contextMenuItems,
    handleDeleteConfirm,
    setCapabilityToDelete,
  } = useCapabilityContextMenu({
    capabilities,
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

  const handleDomainSelect = useCallback((id: BusinessDomainId | null) => {
    setSelectedDomainId(id);
    setSelectedCapability(null);
  }, []);

  const closeCapabilityDetails = useCallback(() => {
    setSelectedCapability(null);
  }, []);

  const closeDeleteDialog = useCallback(() => {
    setCapabilityToDelete(null);
  }, [setCapabilityToDelete]);

  return {
    domains,
    domainsLoading,
    selectedDomainId,
    selectedDomain,
    handleDomainSelect,
    capabilities,
    capabilitiesLoading,
    allCapabilities,
    assignedCapabilityIds,
    treeLoading,
    depth,
    setDepth,
    positions,
    dragHandlers,
    selectedCapabilities,
    handleCapabilityClick,
    handleCapabilityContextMenu,
    selectedCapability,
    closeCapabilityDetails,
    contextMenu,
    contextMenuItems,
    closeContextMenu,
    capabilityToDelete,
    capabilitiesToDelete,
    handleDeleteConfirm,
    closeDeleteDialog,
  };
}
