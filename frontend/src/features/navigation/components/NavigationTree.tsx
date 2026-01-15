import React, { useState } from 'react';
import { useAppStore } from '../../../store/appStore';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { useCapabilities } from '../../capabilities/hooks/useCapabilities';
import { useComponents } from '../../components/hooks/useComponents';
import { useViews } from '../../views/hooks/useViews';
import { DeleteCapabilityDialog } from '../../capabilities/components/DeleteCapabilityDialog';
import { useNavigationTreeState } from '../hooks/useNavigationTreeState';
import { useTreeContextMenus } from '../hooks/useTreeContextMenus';
import { ApplicationsSection } from './sections/ApplicationsSection';
import { ViewsSection } from './sections/ViewsSection';
import { CapabilitiesSection } from './sections/CapabilitiesSection';
import { TreeContextMenus } from './TreeContextMenus';
import { CreateViewDialog } from './CreateViewDialog';
import { DeleteConfirmation } from './DeleteConfirmation';
import type { NavigationTreeProps } from '../types';

export const NavigationTree: React.FC<NavigationTreeProps> = ({
  onComponentSelect,
  onViewSelect,
  onAddComponent,
  onCapabilitySelect,
  onAddCapability,
  onEditCapability,
  onEditComponent,
  canCreateView = true,
}) => {
  const { data: components = [] } = useComponents();
  const { currentView } = useCurrentView();
  const selectedNodeId = useAppStore((state) => state.selectedNodeId);
  const { data: capabilities = [] } = useCapabilities();
  const { data: views = [] } = useViews();

  const [selectedCapabilityId, setSelectedCapabilityId] = useState<string | null>(null);

  const treeState = useNavigationTreeState();
  const contextMenus = useTreeContextMenus({ components, onEditCapability, onEditComponent });

  return (
    <>
      <div className={`navigation-tree ${treeState.isOpen ? 'open' : 'closed'}`}>
        {treeState.isOpen && (
          <div className="navigation-tree-content">
            <div className="navigation-tree-header">
              <h3>Explorer</h3>
              <button
                className="tree-toggle-btn"
                onClick={() => treeState.setIsOpen(false)}
                aria-label="Close navigation"
              >
                ‹
              </button>
            </div>

            <ApplicationsSection
              components={components}
              currentView={currentView}
              selectedNodeId={selectedNodeId}
              isExpanded={treeState.isModelsExpanded}
              onToggle={() => treeState.setIsModelsExpanded(!treeState.isModelsExpanded)}
              onAddComponent={onAddComponent}
              onComponentSelect={onComponentSelect}
              onComponentContextMenu={contextMenus.handleComponentContextMenu}
              editingState={contextMenus.editingState}
              setEditingState={contextMenus.setEditingState}
              onRenameSubmit={contextMenus.handleRenameSubmit}
              editInputRef={contextMenus.editInputRef}
            />

            <ViewsSection
              views={views}
              currentView={currentView}
              isExpanded={treeState.isViewsExpanded}
              onToggle={() => treeState.setIsViewsExpanded(!treeState.isViewsExpanded)}
              canCreateView={canCreateView}
              onCreateView={() => contextMenus.setShowCreateDialog(true)}
              onViewSelect={onViewSelect}
              onViewContextMenu={contextMenus.handleViewContextMenu}
              editingState={contextMenus.editingState}
              setEditingState={contextMenus.setEditingState}
              onRenameSubmit={contextMenus.handleRenameSubmit}
              editInputRef={contextMenus.editInputRef}
            />

            <CapabilitiesSection
              capabilities={capabilities}
              currentView={currentView}
              isExpanded={treeState.isCapabilitiesExpanded}
              onToggle={() => treeState.setIsCapabilitiesExpanded(!treeState.isCapabilitiesExpanded)}
              onAddCapability={onAddCapability}
              onCapabilitySelect={onCapabilitySelect}
              onCapabilityContextMenu={contextMenus.handleCapabilityContextMenu}
              expandedCapabilities={treeState.expandedCapabilities}
              toggleCapabilityExpanded={treeState.toggleCapabilityExpanded}
              selectedCapabilityId={selectedCapabilityId}
              setSelectedCapabilityId={setSelectedCapabilityId}
            />
          </div>
        )}
      </div>

      {!treeState.isOpen && (
        <button
          className="tree-toggle-btn-collapsed"
          onClick={() => treeState.setIsOpen(true)}
          aria-label="Open navigation"
        >
          ›
        </button>
      )}

      <TreeContextMenus
        viewContextMenu={contextMenus.viewContextMenu}
        componentContextMenu={contextMenus.componentContextMenu}
        capabilityContextMenu={contextMenus.capabilityContextMenu}
        getViewContextMenuItems={contextMenus.getViewContextMenuItems}
        getComponentContextMenuItems={contextMenus.getComponentContextMenuItems}
        getCapabilityContextMenuItems={contextMenus.getCapabilityContextMenuItems}
        setViewContextMenu={contextMenus.setViewContextMenu}
        setComponentContextMenu={contextMenus.setComponentContextMenu}
        setCapabilityContextMenu={contextMenus.setCapabilityContextMenu}
      />

      <CreateViewDialog
        isOpen={contextMenus.showCreateDialog}
        viewName={contextMenus.createViewName}
        onViewNameChange={contextMenus.setCreateViewName}
        onClose={() => contextMenus.setShowCreateDialog(false)}
        onCreate={contextMenus.handleCreateView}
      />

      <DeleteConfirmation
        deleteTarget={contextMenus.deleteTarget}
        onConfirm={contextMenus.handleDeleteConfirm}
        onCancel={() => contextMenus.setDeleteTarget(null)}
        isLoading={contextMenus.isDeleting}
      />

      <DeleteCapabilityDialog
        isOpen={contextMenus.deleteCapability !== null}
        onClose={() => contextMenus.setDeleteCapability(null)}
        capability={contextMenus.deleteCapability}
      />
    </>
  );
};
