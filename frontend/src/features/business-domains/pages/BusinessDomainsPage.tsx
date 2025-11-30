import { DndContext } from '@dnd-kit/core';
import { DomainsSidebar } from '../components/DomainsSidebar';
import { CapabilityExplorerSidebar } from '../components/CapabilityExplorerSidebar';
import { VisualizationArea } from '../components/VisualizationArea';
import { CapabilityDetailSidebar } from '../components/CapabilityDetailSidebar';
import { DomainDialogs } from '../components/DomainDialogs';
import { DragOverlayContent } from '../components/DragOverlayContent';
import { PageLoadingStates } from '../components/PageLoadingStates';
import { ContextMenu } from '../../../components/shared/ContextMenu';
import { useBusinessDomainsPage } from '../hooks/useBusinessDomainsPage';
import '../components/visualization.css';

export function BusinessDomainsPage() {
  const {
    domains,
    isLoading,
    error,
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
  } = useBusinessDomainsPage();

  return (
    <PageLoadingStates isLoading={isLoading} hasData={domains.length > 0} error={error}>
      <DndContext sensors={sensors} onDragStart={dragHandlers.handleDragStart} onDragEnd={dragHandlers.handleDragEnd}>
        <div className="business-domains-layout" data-testid="business-domains-page" style={{ display: 'flex', height: '100vh', position: 'relative' }}>
          <DomainsSidebar
            isCollapsed={sidebarState.isDomainsSidebarCollapsed}
            domains={domains}
            selectedDomainId={visualizedDomain?.id}
            onToggle={() => sidebarState.setIsDomainsSidebarCollapsed(!sidebarState.isDomainsSidebarCollapsed)}
            onCreateClick={dialogManager.handleCreateClick}
            onVisualize={handleVisualizeClick}
            onContextMenu={contextMenu.handleContextMenu}
          />

          <VisualizationArea
            visualizedDomain={visualizedDomain}
            capabilities={capabilities}
            capabilitiesWithDescendants={filtering.capabilitiesWithDescendants}
            capabilitiesLoading={capabilitiesLoading}
            depth={depth}
            positions={positions}
            onDepthChange={setDepth}
            onCapabilityClick={handleCapabilityClick}
          />

          <CapabilityExplorerSidebar
            isCollapsed={sidebarState.isExplorerSidebarCollapsed}
            visualizedDomain={visualizedDomain}
            capabilities={filtering.allCapabilities}
            assignedCapabilityIds={filtering.assignedCapabilityIds}
            isLoading={false}
            onToggle={() => sidebarState.setIsExplorerSidebarCollapsed(!sidebarState.isExplorerSidebarCollapsed)}
          />

          <CapabilityDetailSidebar
            capability={selectedCapability}
            onClose={() => handleCapabilityClick(null)}
          />
        </div>

        <DragOverlayContent activeCapability={dragHandlers.activeCapability} />

        {contextMenu.contextMenu && (
          <ContextMenu
            x={contextMenu.contextMenu.x}
            y={contextMenu.contextMenu.y}
            items={contextMenu.getContextMenuItems(contextMenu.contextMenu)}
            onClose={contextMenu.closeContextMenu}
          />
        )}

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

      </DndContext>
    </PageLoadingStates>
  );
}
