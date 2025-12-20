import { DockviewBusinessDomainsLayout } from '../components/DockviewBusinessDomainsLayout';
import { DomainDialogs } from '../components/DomainDialogs';
import { PageLoadingStates } from '../components/PageLoadingStates';
import { ContextMenu } from '../../../components/shared/ContextMenu';
import { DeleteCapabilityDialog } from '../../capabilities/components/DeleteCapabilityDialog';
import { useBusinessDomainsPage } from '../hooks/useBusinessDomainsPage';
import '../components/visualization.css';

export function BusinessDomainsPage() {
  const hookData = useBusinessDomainsPage();
  const {
    domains,
    isLoading,
    error,
    dialogManager,
    domainContextMenu,
    capabilityContextMenu,
  } = hookData;

  return (
    <PageLoadingStates isLoading={isLoading} hasData={domains.length > 0} error={error}>
      <DockviewBusinessDomainsLayout hookData={hookData} />

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
