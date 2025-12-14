import { useState, useCallback } from 'react';
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
import { useApplicationSettings } from './useApplicationSettings';
import { useCapabilityRealizations } from './useCapabilityRealizations';
import type { BusinessDomain, BusinessDomainId, Capability, ComponentId } from '../../../api/types';

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

  const { getRealizationsForCapability, refetch: refetchRealizations } = useCapabilityRealizations(
    showApplications,
    visualizedDomain?.id as BusinessDomainId | null,
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
    setSelectedComponentId(null);
  };

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
    sensors,
    capabilities,
    capabilitiesLoading,
    filtering,
    dragHandlers,
    contextMenu,
    handleVisualizeClick,
    handleCapabilityClick,
    getRealizationsForCapability,
    handleApplicationClick,
    clearSelectedComponent,
  };
}
