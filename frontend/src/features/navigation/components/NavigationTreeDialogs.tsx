import React from 'react';
import type { Capability } from '../../../api/types';
import { DeleteCapabilityDialog } from '../../capabilities/components/DeleteCapabilityDialog';
import { InviteToEditDialog } from '../../edit-grants/components/InviteToEditDialog';
import { useCreateEditGrant } from '../../edit-grants/hooks/useEditGrants';
import { CreateAcquiredEntityDialog } from '../../origin-entities/components/CreateAcquiredEntityDialog';
import { CreateVendorDialog } from '../../origin-entities/components/CreateVendorDialog';
import { CreateInternalTeamDialog } from '../../origin-entities/components/CreateInternalTeamDialog';
import { TreeContextMenus } from './TreeContextMenus';
import { CreateViewDialog } from './CreateViewDialog';
import { DeleteConfirmation } from './DeleteConfirmation';
import { TreeMultiSelectContextMenu } from './TreeMultiSelectContextMenu';
import { ConfirmationDialog } from '../../../components/shared/ConfirmationDialog';
import type { DeleteTarget } from './DeleteConfirmation';
import type { ViewContextMenuState, ComponentContextMenuState, CapabilityContextMenuState } from '../types';
import type { OriginEntityContextMenuState, InviteTarget } from '../hooks/useTreeContextMenus';
import type { ContextMenuItem } from '../../../components/shared/ContextMenu';
import type { TreeMultiSelectMenuState } from '../hooks/useTreeMultiSelectMenu';
import type { TreeBulkOperationRequest } from './TreeMultiSelectContextMenu';
import type { TreeBulkOperationResult } from '../hooks/useTreeBulkDelete';
import type { TreeSelectedItem } from '../hooks/useTreeMultiSelect';

type OriginEntityDialogType = 'acquired' | 'vendor' | 'team' | null;

function buildBulkDeleteError(result: TreeBulkOperationResult): string {
  const failedNames = result.failed.map((f) => f.name).join(', ');
  return `${result.succeeded.length} item(s) succeeded, ${result.failed.length} failed: ${failedNames}`;
}

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
    inviteTarget: InviteTarget | null;
    setInviteTarget: (v: InviteTarget | null) => void;
  };
  openOriginDialog: OriginEntityDialogType;
  onCloseOriginDialog: () => void;
  multiSelectMenu: TreeMultiSelectMenuState | null;
  onCloseMultiSelectMenu: () => void;
  onRequestBulkOperation: (request: TreeBulkOperationRequest) => void;
  bulkDelete: {
    bulkItems: TreeSelectedItem[] | null;
    isExecuting: boolean;
    result: TreeBulkOperationResult | null;
    itemNames: string[];
    handleCancel: () => void;
  };
  onBulkDeleteConfirm: () => void;
}

export const NavigationTreeDialogs: React.FC<NavigationTreeDialogsProps> = ({
  contextMenus,
  openOriginDialog,
  onCloseOriginDialog,
  multiSelectMenu,
  onCloseMultiSelectMenu,
  onRequestBulkOperation,
  bulkDelete,
  onBulkDeleteConfirm,
}) => {
  const createGrant = useCreateEditGrant();

  return (
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

    <TreeMultiSelectContextMenu
      menu={multiSelectMenu}
      onClose={onCloseMultiSelectMenu}
      onRequestBulkOperation={onRequestBulkOperation}
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

    {bulkDelete.bulkItems && (
      <ConfirmationDialog
        title={`Delete ${bulkDelete.bulkItems.length} items from Model`}
        message={`This will permanently delete ${bulkDelete.bulkItems.length} items from the entire model. They will be removed from ALL views and all associated relations will be deleted. This cannot be undone.`}
        itemNames={bulkDelete.itemNames}
        confirmText={`Delete ${bulkDelete.bulkItems.length} items`}
        cancelText="Cancel"
        onConfirm={onBulkDeleteConfirm}
        onCancel={bulkDelete.handleCancel}
        isLoading={bulkDelete.isExecuting}
        error={bulkDelete.result ? buildBulkDeleteError(bulkDelete.result) : null}
      />
    )}

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

    {contextMenus.inviteTarget && (
      <InviteToEditDialog
        isOpen={contextMenus.inviteTarget !== null}
        onClose={() => contextMenus.setInviteTarget(null)}
        onSubmit={async (request) => { await createGrant.mutateAsync(request); }}
        artifactType={contextMenus.inviteTarget.artifactType}
        artifactId={contextMenus.inviteTarget.id}
      />
    )}
  </>
  );
};
