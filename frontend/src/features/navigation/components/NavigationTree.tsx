import React, { useState, useMemo } from 'react';
import { useAppStore } from '../../../store/appStore';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { useCapabilities } from '../../capabilities/hooks/useCapabilities';
import { useComponents } from '../../components/hooks/useComponents';
import { useViews } from '../../views/hooks/useViews';
import { useAcquiredEntitiesQuery } from '../../origin-entities/hooks/useAcquiredEntities';
import { useVendorsQuery } from '../../origin-entities/hooks/useVendors';
import { useInternalTeamsQuery } from '../../origin-entities/hooks/useInternalTeams';
import { useOriginRelationshipsQuery } from '../../origin-entities/hooks/useOriginRelationships';
import { useNavigationTreeState } from '../hooks/useNavigationTreeState';
import { useTreeContextMenus } from '../hooks/useTreeContextMenus';
import { NavigationTreeContent } from './NavigationTreeContent';
import { NavigationTreeDialogs } from './NavigationTreeDialogs';
import type { NavigationTreeProps } from '../types';

type OriginEntityDialogType = 'acquired' | 'vendor' | 'team' | null;

interface SelectedEntityIds {
  acquiredEntityId: string | null;
  vendorId: string | null;
  teamId: string | null;
}

function extractSelectedEntityIds(nodeId: string | null): SelectedEntityIds {
  if (!nodeId) {
    return { acquiredEntityId: null, vendorId: null, teamId: null };
  }
  return {
    acquiredEntityId: nodeId.startsWith('acq-') ? nodeId.slice(4) : null,
    vendorId: nodeId.startsWith('vendor-') ? nodeId.slice(7) : null,
    teamId: nodeId.startsWith('team-') ? nodeId.slice(5) : null,
  };
}

export const NavigationTree: React.FC<NavigationTreeProps> = ({
  onComponentSelect,
  onViewSelect,
  onAddComponent,
  onCapabilitySelect,
  onAddCapability,
  onEditCapability,
  onEditComponent,
  onOriginEntitySelect,
  canCreateView = true,
  canCreateOriginEntity = false,
}) => {
  const { data: components = [] } = useComponents();
  const { currentView } = useCurrentView();
  const selectedNodeId = useAppStore((state) => state.selectedNodeId);
  const { data: capabilities = [] } = useCapabilities();
  const { data: views = [] } = useViews();
  const { data: acquiredEntities = [] } = useAcquiredEntitiesQuery();
  const { data: vendors = [] } = useVendorsQuery();
  const { data: internalTeams = [] } = useInternalTeamsQuery();
  const { data: originRelationships = [] } = useOriginRelationshipsQuery();

  const [selectedCapabilityId, setSelectedCapabilityId] = useState<string | null>(null);
  const [openOriginDialog, setOpenOriginDialog] = useState<OriginEntityDialogType>(null);

  const treeState = useNavigationTreeState();
  const contextMenus = useTreeContextMenus({ components, onEditCapability, onEditComponent });

  const selectedEntityIds = useMemo(
    () => extractSelectedEntityIds(selectedNodeId),
    [selectedNodeId]
  );

  const originEntityAddHandler = canCreateOriginEntity
    ? (type: OriginEntityDialogType) => setOpenOriginDialog(type)
    : undefined;

  return (
    <>
      <div className={`navigation-tree ${treeState.isOpen ? 'open' : 'closed'}`}>
        {treeState.isOpen && (
          <NavigationTreeContent
            components={components}
            currentView={currentView}
            selectedNodeId={selectedNodeId}
            capabilities={capabilities}
            views={views}
            acquiredEntities={acquiredEntities}
            vendors={vendors}
            internalTeams={internalTeams}
            originRelationships={originRelationships}
            selectedCapabilityId={selectedCapabilityId}
            setSelectedCapabilityId={setSelectedCapabilityId}
            selectedEntityIds={selectedEntityIds}
            treeState={treeState}
            contextMenus={contextMenus}
            onComponentSelect={onComponentSelect}
            onViewSelect={onViewSelect}
            onAddComponent={onAddComponent}
            onCapabilitySelect={onCapabilitySelect}
            onAddCapability={onAddCapability}
            onOriginEntitySelect={onOriginEntitySelect}
            canCreateView={canCreateView}
            onAddAcquiredEntity={originEntityAddHandler ? () => originEntityAddHandler('acquired') : undefined}
            onAddVendor={originEntityAddHandler ? () => originEntityAddHandler('vendor') : undefined}
            onAddTeam={originEntityAddHandler ? () => originEntityAddHandler('team') : undefined}
          />
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

      <NavigationTreeDialogs
        contextMenus={contextMenus}
        openOriginDialog={openOriginDialog}
        onCloseOriginDialog={() => setOpenOriginDialog(null)}
      />
    </>
  );
};
