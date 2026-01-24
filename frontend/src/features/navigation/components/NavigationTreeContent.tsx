import React from 'react';
import type { Component, Capability, View, AcquiredEntity, Vendor, InternalTeam } from '../../../api/types';
import { ApplicationsSection } from './sections/ApplicationsSection';
import { ViewsSection } from './sections/ViewsSection';
import { CapabilitiesSection } from './sections/CapabilitiesSection';
import { AcquiredEntitiesSection } from './sections/AcquiredEntitiesSection';
import { VendorsSection } from './sections/VendorsSection';
import { InternalTeamsSection } from './sections/InternalTeamsSection';
import type { EditingState } from '../types';

interface SelectedEntityIds {
  acquiredEntityId: string | null;
  vendorId: string | null;
  teamId: string | null;
}

interface NavigationTreeContentProps {
  components: Component[];
  currentView: View | null;
  selectedNodeId: string | null;
  capabilities: Capability[];
  views: View[];
  acquiredEntities: AcquiredEntity[];
  vendors: Vendor[];
  internalTeams: InternalTeam[];
  selectedCapabilityId: string | null;
  setSelectedCapabilityId: (id: string | null) => void;
  selectedEntityIds: SelectedEntityIds;
  treeState: {
    isModelsExpanded: boolean;
    setIsModelsExpanded: (v: boolean) => void;
    isViewsExpanded: boolean;
    setIsViewsExpanded: (v: boolean) => void;
    isCapabilitiesExpanded: boolean;
    setIsCapabilitiesExpanded: (v: boolean) => void;
    isAcquiredEntitiesExpanded: boolean;
    setIsAcquiredEntitiesExpanded: (v: boolean) => void;
    isVendorsExpanded: boolean;
    setIsVendorsExpanded: (v: boolean) => void;
    isInternalTeamsExpanded: boolean;
    setIsInternalTeamsExpanded: (v: boolean) => void;
    expandedCapabilities: Set<string>;
    toggleCapabilityExpanded: (id: string) => void;
    setIsOpen: (v: boolean) => void;
  };
  contextMenus: {
    handleComponentContextMenu: (e: React.MouseEvent, component: Component) => void;
    handleViewContextMenu: (e: React.MouseEvent, view: View) => void;
    handleCapabilityContextMenu: (e: React.MouseEvent, capability: Capability) => void;
    handleAcquiredEntityContextMenu: (e: React.MouseEvent, entity: AcquiredEntity) => void;
    handleVendorContextMenu: (e: React.MouseEvent, vendor: Vendor) => void;
    handleInternalTeamContextMenu: (e: React.MouseEvent, team: InternalTeam) => void;
    editingState: EditingState | null;
    setEditingState: (state: EditingState | null) => void;
    handleRenameSubmit: () => void;
    editInputRef: React.RefObject<HTMLInputElement | null>;
    setShowCreateDialog: (v: boolean) => void;
  };
  onComponentSelect?: (componentId: string) => void;
  onViewSelect?: (viewId: string) => void;
  onAddComponent?: () => void;
  onCapabilitySelect?: (capabilityId: string) => void;
  onAddCapability?: () => void;
  onOriginEntitySelect?: (nodeId: string) => void;
  canCreateView: boolean;
  onAddAcquiredEntity?: () => void;
  onAddVendor?: () => void;
  onAddTeam?: () => void;
}

export const NavigationTreeContent: React.FC<NavigationTreeContentProps> = ({
  components,
  currentView,
  selectedNodeId,
  capabilities,
  views,
  acquiredEntities,
  vendors,
  internalTeams,
  selectedCapabilityId,
  setSelectedCapabilityId,
  selectedEntityIds,
  treeState,
  contextMenus,
  onComponentSelect,
  onViewSelect,
  onAddComponent,
  onCapabilitySelect,
  onAddCapability,
  onOriginEntitySelect,
  canCreateView,
  onAddAcquiredEntity,
  onAddVendor,
  onAddTeam,
}) => (
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

    <AcquiredEntitiesSection
      acquiredEntities={acquiredEntities}
      currentView={currentView}
      selectedEntityId={selectedEntityIds.acquiredEntityId}
      isExpanded={treeState.isAcquiredEntitiesExpanded}
      onToggle={() => treeState.setIsAcquiredEntitiesExpanded(!treeState.isAcquiredEntitiesExpanded)}
      onAddEntity={onAddAcquiredEntity}
      onEntitySelect={(entityId) => onOriginEntitySelect?.(`acq-${entityId}`)}
      onEntityContextMenu={contextMenus.handleAcquiredEntityContextMenu}
    />

    <VendorsSection
      vendors={vendors}
      currentView={currentView}
      selectedVendorId={selectedEntityIds.vendorId}
      isExpanded={treeState.isVendorsExpanded}
      onToggle={() => treeState.setIsVendorsExpanded(!treeState.isVendorsExpanded)}
      onAddVendor={onAddVendor}
      onVendorSelect={(vendorId) => onOriginEntitySelect?.(`vendor-${vendorId}`)}
      onVendorContextMenu={contextMenus.handleVendorContextMenu}
    />

    <InternalTeamsSection
      internalTeams={internalTeams}
      currentView={currentView}
      selectedTeamId={selectedEntityIds.teamId}
      isExpanded={treeState.isInternalTeamsExpanded}
      onToggle={() => treeState.setIsInternalTeamsExpanded(!treeState.isInternalTeamsExpanded)}
      onAddTeam={onAddTeam}
      onTeamSelect={(teamId) => onOriginEntitySelect?.(`team-${teamId}`)}
      onTeamContextMenu={contextMenus.handleInternalTeamContextMenu}
    />
  </div>
);
