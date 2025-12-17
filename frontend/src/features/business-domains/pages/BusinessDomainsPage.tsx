import { DomainsSidebar } from '../components/DomainsSidebar';
import { CapabilityExplorerSidebar } from '../components/CapabilityExplorerSidebar';
import { VisualizationArea } from '../components/VisualizationArea';
import { DetailsSidebar } from '../components/DetailsSidebar';
import { DomainDialogs } from '../components/DomainDialogs';
import { PageLoadingStates } from '../components/PageLoadingStates';
import { ContextMenu } from '../../../components/shared/ContextMenu';
import { DeleteCapabilityDialog } from '../../capabilities/components/DeleteCapabilityDialog';
import { useBusinessDomainsPage } from '../hooks/useBusinessDomainsPage';
import '../components/visualization.css';

export function BusinessDomainsPage() {
  const {
    domains,
    isLoading,
    error,
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
  } = useBusinessDomainsPage();

  return (
    <PageLoadingStates isLoading={isLoading} hasData={domains.length > 0} error={error}>
      <div className="business-domains-layout" data-testid="business-domains-page" style={{ display: 'flex', height: '100vh', position: 'relative' }}>
        <DomainsSidebar
          isCollapsed={sidebarState.isDomainsSidebarCollapsed}
          domains={domains}
          selectedDomainId={visualizedDomain?.id}
          onToggle={() => sidebarState.setIsDomainsSidebarCollapsed(!sidebarState.isDomainsSidebarCollapsed)}
          onCreateClick={dialogManager.handleCreateClick}
          onVisualize={handleVisualizeClick}
          onContextMenu={domainContextMenu.handleContextMenu}
        />

        <VisualizationArea
          visualizedDomain={visualizedDomain}
          capabilities={filtering.capabilitiesWithDescendants}
          capabilitiesLoading={capabilitiesLoading}
          depth={depth}
          positions={positions}
          onDepthChange={setDepth}
          onCapabilityClick={handleCapabilityClick}
          onContextMenu={capabilityContextMenu.handleCapabilityContextMenu}
          selectedCapabilities={selectedCapabilities}
          showApplications={showApplications}
          onShowApplicationsChange={setShowApplications}
          getRealizationsForCapability={getRealizationsForCapability}
          onApplicationClick={handleApplicationClick}
          isDragOver={dragHandlers.isDragOver}
          onDragOver={dragHandlers.handleDragOver}
          onDragLeave={dragHandlers.handleDragLeave}
          onDrop={dragHandlers.handleDrop}
        />

        <CapabilityExplorerSidebar
          isCollapsed={sidebarState.isExplorerSidebarCollapsed}
          visualizedDomain={visualizedDomain}
          capabilities={filtering.allCapabilities}
          assignedCapabilityIds={filtering.assignedCapabilityIds}
          isLoading={false}
          onToggle={() => sidebarState.setIsExplorerSidebarCollapsed(!sidebarState.isExplorerSidebarCollapsed)}
          onDragStart={dragHandlers.handleDragStart}
          onDragEnd={dragHandlers.handleDragEnd}
        />

        <DetailsSidebar
          selectedCapability={selectedCapability}
          selectedComponentId={selectedComponentId}
          onCloseCapability={clearCapabilityDetails}
          onCloseApplication={clearSelectedComponent}
        />
      </div>

      {domainContextMenu.contextMenu && (
        <ContextMenu
          x={domainContextMenu.contextMenu.x}
          y={domainContextMenu.contextMenu.y}
          items={domainContextMenu.getContextMenuItems(domainContextMenu.contextMenu)}
          onClose={domainContextMenu.closeContextMenu}
        />
      )}

      {capabilityContextMenu.contextMenu && (
        <ContextMenu
          x={capabilityContextMenu.contextMenu.x}
          y={capabilityContextMenu.contextMenu.y}
          items={capabilityContextMenu.contextMenuItems}
          onClose={capabilityContextMenu.closeContextMenu}
        />
      )}

      <DeleteCapabilityDialog
        isOpen={capabilityContextMenu.capabilityToDelete !== null}
        onClose={() => capabilityContextMenu.setCapabilityToDelete(null)}
        capability={capabilityContextMenu.capabilityToDelete}
        onConfirm={capabilityContextMenu.handleDeleteConfirm}
        capabilitiesToDelete={capabilityContextMenu.capabilitiesToDelete}
      />

      <DomainDialogs
        dialogMode={dialogManager.dialogMode}
        selectedDomain={dialogManager.selectedDomain}
        domainToDelete={dialogManager.domainToDelete}
        dialogRef={dialogManager.dialogRef}
        onFormSubmit={dialogManager.handleFormSubmit}
        onFormCancel={dialogManager.handleFormCancel}
        onConfirmDelete={dialogManager.handleConfirmDelete}
        onCancelDelete={dialogManager.handleCancelDelete}
      />
    </PageLoadingStates>
  );
}
