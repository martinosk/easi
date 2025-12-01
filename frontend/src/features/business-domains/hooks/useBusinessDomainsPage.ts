import { useState } from 'react';
import { useSensor, useSensors, PointerSensor } from '@dnd-kit/core';
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
import type { BusinessDomain, Capability } from '../../../api/types';

export function useBusinessDomainsPage() {
  const [visualizedDomain, setVisualizedDomain] = useState<BusinessDomain | null>(null);
  const [selectedCapability, setSelectedCapability] = useState<Capability | null>(null);
  const [depth, setDepth] = usePersistedDepth();

  const { domains, isLoading, error, createDomain, updateDomain, deleteDomain } = useBusinessDomains();
  const { tree, isLoading: treeLoading } = useCapabilityTree();

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

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 8,
      },
    })
  );

  const capabilitiesLink = visualizedDomain?._links.capabilities;

  const {
    capabilities,
    isLoading: capabilitiesLoading,
    associateCapability,
    refetch: refetchCapabilities,
  } = useDomainCapabilities(capabilitiesLink);

  const filtering = useCapabilityFiltering(tree, capabilities);

  const dragHandlers = useDragHandlers({
    domainId: visualizedDomain?.id ?? null,
    capabilities,
    assignedCapabilityIds: filtering.assignedCapabilityIds,
    positions,
    updatePosition,
    associateCapability,
    refetchCapabilities,
  });

  const contextMenu = useDomainContextMenu({
    onEdit: dialogManager.handleEditClick,
    onDelete: dialogManager.handleDeleteClick,
  });

  const handleVisualizeClick = (domain: BusinessDomain) => {
    setVisualizedDomain(domain);
    setSelectedCapability(null);
  };

  const handleCapabilityClick = (capability: Capability | null) => {
    setSelectedCapability(capability);
  };

  return {
    domains,
    isLoading,
    error,
    tree,
    treeLoading,
    visualizedDomain,
    selectedCapability,
    depth,
    setDepth,
    sidebarState,
    dialogManager,
    positions,
    sensors,
    capabilities,
    capabilitiesLoading,
    filtering,
    dragHandlers,
    contextMenu,
    handleVisualizeClick,
    handleCapabilityClick,
  };
}
