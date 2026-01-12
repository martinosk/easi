import React, { useState } from 'react';
import { useAppStore } from '../../../store/appStore';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { useCapabilities } from '../../capabilities/hooks/useCapabilities';
import { useComponents } from '../../components/hooks/useComponents';
import { useViews } from '../../views/hooks/useViews';
import { ContextMenu } from '../../../components/shared/ContextMenu';
import { ConfirmationDialog } from '../../../components/shared/ConfirmationDialog';
import { DeleteCapabilityDialog } from '../../capabilities/components/DeleteCapabilityDialog';
import { useNavigationTreeState } from '../hooks/useNavigationTreeState';
import { useTreeContextMenus } from '../hooks/useTreeContextMenus';
import { ApplicationsSection } from './sections/ApplicationsSection';
import { ViewsSection } from './sections/ViewsSection';
import { CapabilitiesSection } from './sections/CapabilitiesSection';
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

      {contextMenus.viewContextMenu && (
        <ContextMenu
          x={contextMenus.viewContextMenu.x}
          y={contextMenus.viewContextMenu.y}
          items={contextMenus.getViewContextMenuItems(contextMenus.viewContextMenu)}
          onClose={() => contextMenus.setViewContextMenu(null)}
        />
      )}

      {contextMenus.componentContextMenu && (
        <ContextMenu
          x={contextMenus.componentContextMenu.x}
          y={contextMenus.componentContextMenu.y}
          items={contextMenus.getComponentContextMenuItems(contextMenus.componentContextMenu)}
          onClose={() => contextMenus.setComponentContextMenu(null)}
        />
      )}

      {contextMenus.capabilityContextMenu && (
        <ContextMenu
          x={contextMenus.capabilityContextMenu.x}
          y={contextMenus.capabilityContextMenu.y}
          items={contextMenus.getCapabilityContextMenuItems(contextMenus.capabilityContextMenu)}
          onClose={() => contextMenus.setCapabilityContextMenu(null)}
        />
      )}

      {contextMenus.showCreateDialog && (
        <div className="dialog-overlay" onClick={() => contextMenus.setShowCreateDialog(false)}>
          <div className="dialog" onClick={(e) => e.stopPropagation()}>
            <h3>Create New View</h3>
            <input
              type="text"
              placeholder="View name"
              value={contextMenus.createViewName}
              onChange={(e) => contextMenus.setCreateViewName(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') contextMenus.handleCreateView();
                if (e.key === 'Escape') contextMenus.setShowCreateDialog(false);
              }}
              autoFocus
              className="dialog-input"
            />
            <div className="dialog-actions">
              <button onClick={() => contextMenus.setShowCreateDialog(false)} className="btn-secondary">
                Cancel
              </button>
              <button onClick={contextMenus.handleCreateView} className="btn-primary">
                Create
              </button>
            </div>
          </div>
        </div>
      )}

      {contextMenus.deleteTarget && (
        <ConfirmationDialog
          title={contextMenus.deleteTarget.type === 'view' ? 'Delete View' : 'Delete Application'}
          message={
            contextMenus.deleteTarget.type === 'view'
              ? `Are you sure you want to delete this view?`
              : `This will delete the application from the entire model, remove it from ALL views, and delete ALL relations involving this application.`
          }
          itemName={contextMenus.deleteTarget.type === 'view' ? contextMenus.deleteTarget.view!.name : contextMenus.deleteTarget.component!.name}
          confirmText="Delete"
          cancelText="Cancel"
          onConfirm={contextMenus.handleDeleteConfirm}
          onCancel={() => contextMenus.setDeleteTarget(null)}
          isLoading={contextMenus.isDeleting}
        />
      )}

      <DeleteCapabilityDialog
        isOpen={contextMenus.deleteCapability !== null}
        onClose={() => contextMenus.setDeleteCapability(null)}
        capability={contextMenus.deleteCapability}
      />
    </>
  );
};
