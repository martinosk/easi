import React from 'react';
import type { Capability } from '../../../api/types';
import { DeleteCapabilityDialog } from '../../capabilities/components/DeleteCapabilityDialog';
import { CreateAcquiredEntityDialog } from '../../origin-entities/components/CreateAcquiredEntityDialog';
import { CreateVendorDialog } from '../../origin-entities/components/CreateVendorDialog';
import { CreateInternalTeamDialog } from '../../origin-entities/components/CreateInternalTeamDialog';
import { TreeContextMenus } from './TreeContextMenus';
import { CreateViewDialog } from './CreateViewDialog';
import { DeleteConfirmation } from './DeleteConfirmation';
import type { DeleteTarget } from './DeleteConfirmation';
import type { ViewContextMenuState, ComponentContextMenuState, CapabilityContextMenuState } from '../types';
import type { OriginEntityContextMenuState } from '../hooks/useTreeContextMenus';
import type { ContextMenuItem } from '../../../components/shared/ContextMenu';

type OriginEntityDialogType = 'acquired' | 'vendor' | 'team' | null;

interface NavigationTreeDialogsProps {
  contextMenus: {
    viewContextMenu: ViewContextMenuState | null;
    setViewContextMenu: (v: ViewContextMenuState | null) => void;
    componentContextMenu: ComponentContextMenuState | null;
    setComponentContextMenu: (v: ComponentContextMenuState | null) => void;
    capabilityContextMenu: CapabilityContextMenuState | null;
    setCapabilityContextMenu: (v: CapabilityContextMenuState | null) => void;
    originEntityContextMenu: OriginEntityContextMenuState | null;
    setOriginEntityContextMenu: (v: OriginEntityContextMenuState | null) => void;
    getViewContextMenuItems: (menu: ViewContextMenuState) => ContextMenuItem[];
    getComponentContextMenuItems: (menu: ComponentContextMenuState) => ContextMenuItem[];
    getCapabilityContextMenuItems: (menu: CapabilityContextMenuState) => ContextMenuItem[];
    getOriginEntityContextMenuItems: (menu: OriginEntityContextMenuState) => ContextMenuItem[];
    showCreateDialog: boolean;
    setShowCreateDialog: (v: boolean) => void;
    createViewName: string;
    setCreateViewName: (v: string) => void;
    handleCreateView: () => void;
    deleteTarget: DeleteTarget | null;
    setDeleteTarget: (v: DeleteTarget | null) => void;
    handleDeleteConfirm: () => void;
    isDeleting: boolean;
    deleteCapability: Capability | null;
    setDeleteCapability: (v: Capability | null) => void;
  };
  openOriginDialog: OriginEntityDialogType;
  onCloseOriginDialog: () => void;
}

export const NavigationTreeDialogs: React.FC<NavigationTreeDialogsProps> = ({
  contextMenus,
  openOriginDialog,
  onCloseOriginDialog,
}) => (
  <>
    <TreeContextMenus
      viewContextMenu={contextMenus.viewContextMenu}
      componentContextMenu={contextMenus.componentContextMenu}
      capabilityContextMenu={contextMenus.capabilityContextMenu}
      originEntityContextMenu={contextMenus.originEntityContextMenu}
      getViewContextMenuItems={contextMenus.getViewContextMenuItems}
      getComponentContextMenuItems={contextMenus.getComponentContextMenuItems}
      getCapabilityContextMenuItems={contextMenus.getCapabilityContextMenuItems}
      getOriginEntityContextMenuItems={contextMenus.getOriginEntityContextMenuItems}
      setViewContextMenu={contextMenus.setViewContextMenu}
      setComponentContextMenu={contextMenus.setComponentContextMenu}
      setCapabilityContextMenu={contextMenus.setCapabilityContextMenu}
      setOriginEntityContextMenu={contextMenus.setOriginEntityContextMenu}
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

    <CreateAcquiredEntityDialog
      isOpen={openOriginDialog === 'acquired'}
      onClose={onCloseOriginDialog}
    />

    <CreateVendorDialog
      isOpen={openOriginDialog === 'vendor'}
      onClose={onCloseOriginDialog}
    />

    <CreateInternalTeamDialog
      isOpen={openOriginDialog === 'team'}
      onClose={onCloseOriginDialog}
    />
  </>
);
