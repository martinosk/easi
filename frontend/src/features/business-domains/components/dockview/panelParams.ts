import type { useBusinessDomainsPage } from '../../hooks/useBusinessDomainsPage';

type BusinessDomainsHookReturn = ReturnType<typeof useBusinessDomainsPage>;

export function buildDomainsParams(hookData: BusinessDomainsHookReturn) {
  return {
    domains: hookData.domains,
    selectedDomainId: hookData.visualizedDomain?.id,
    onCreateClick: hookData.dialogManager.handleCreateClick,
    onVisualize: hookData.handleVisualizeClick,
    onContextMenu: hookData.domainContextMenu.handleContextMenu,
  };
}

export function buildVisualizationParams(hookData: BusinessDomainsHookReturn) {
  return {
    visualizedDomain: hookData.visualizedDomain,
    capabilities: hookData.filtering.capabilitiesWithDescendants,
    capabilitiesLoading: hookData.capabilitiesLoading,
    depth: hookData.depth,
    positions: hookData.positions,
    selectedCapabilities: hookData.selectedCapabilities,
    showApplications: hookData.showApplications,
    isDragOver: hookData.dragHandlers.isDragOver,
    onDepthChange: hookData.setDepth,
    onCapabilityClick: hookData.handleCapabilityClick,
    onContextMenu: hookData.capabilityContextMenu.handleCapabilityContextMenu,
    onShowApplicationsChange: hookData.setShowApplications,
    getRealizationsForCapability: hookData.getRealizationsForCapability,
    onApplicationClick: hookData.handleApplicationClick,
    onDragOver: hookData.dragHandlers.handleDragOver,
    onDragLeave: hookData.dragHandlers.handleDragLeave,
    onDrop: hookData.dragHandlers.handleDrop,
  };
}

export function buildExplorerParams(hookData: BusinessDomainsHookReturn) {
  return {
    visualizedDomain: hookData.visualizedDomain,
    capabilities: hookData.filtering.allCapabilities,
    assignedCapabilityIds: hookData.filtering.assignedCapabilityIds,
    onDragStart: hookData.dragHandlers.handleDragStart,
    onDragEnd: hookData.dragHandlers.handleDragEnd,
  };
}

export function buildDetailsParams(hookData: BusinessDomainsHookReturn) {
  return {
    selectedCapability: hookData.selectedCapability,
    selectedComponentId: hookData.selectedComponentId,
    visualizedDomain: hookData.visualizedDomain,
  };
}
