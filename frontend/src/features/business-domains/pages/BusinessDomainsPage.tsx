import { ContextMenu } from '../../../components/shared/ContextMenu';
import { DeleteCapabilityDialog } from '../../capabilities/components/DeleteCapabilityDialog';
import { InviteToEditDialog } from '../../edit-grants/components/InviteToEditDialog';
import { useCreateEditGrant } from '../../edit-grants/hooks/useEditGrants';
import { ApplicationDrawer } from '../components/ApplicationDrawer';
import { CapabilityDrawer } from '../components/CapabilityDrawer';
import { DomainBoard } from '../components/DomainBoard';
import { DomainDialogs } from '../components/DomainDialogs';
import { PageLoadingStates } from '../components/PageLoadingStates';
import { useBusinessDomainsPage } from '../hooks/useBusinessDomainsPage';

export function BusinessDomainsPage() {
  const hookData = useBusinessDomainsPage();
  const { boardDomains, isLoading, error, dialogManager, domainContextMenu, capabilityContextMenu } = hookData;
  const createGrant = useCreateEditGrant();

  return (
    <PageLoadingStates isLoading={isLoading} hasData={boardDomains.length > 0} error={error}>
      <DomainBoard hookData={hookData} />

      {domainContextMenu.contextMenu && (
        <ContextMenu
          x={domainContextMenu.contextMenu.x}
          y={domainContextMenu.contextMenu.y}
          title={domainContextMenu.contextMenu.domain.name}
          items={domainContextMenu.getContextMenuItems(domainContextMenu.contextMenu)}
          onClose={domainContextMenu.closeContextMenu}
        />
      )}

      {capabilityContextMenu.contextMenu && (
        <ContextMenu
          x={capabilityContextMenu.contextMenu.x}
          y={capabilityContextMenu.contextMenu.y}
          title={capabilityContextMenu.contextMenu.capability.name}
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

      {capabilityContextMenu.capabilityToInvite && (
        <InviteToEditDialog
          isOpen={capabilityContextMenu.capabilityToInvite !== null}
          onClose={() => capabilityContextMenu.setCapabilityToInvite(null)}
          onSubmit={async (request) => {
            await createGrant.mutateAsync(request);
          }}
          artifactType="capability"
          artifactId={capabilityContextMenu.capabilityToInvite.id}
        />
      )}

      {domainContextMenu.domainToInvite && (
        <InviteToEditDialog
          isOpen={domainContextMenu.domainToInvite !== null}
          onClose={() => domainContextMenu.setDomainToInvite(null)}
          onSubmit={async (request) => {
            await createGrant.mutateAsync(request);
          }}
          artifactType={domainContextMenu.domainToInvite.artifactType}
          artifactId={domainContextMenu.domainToInvite.id}
        />
      )}

      <DomainDialogs
        dialogMode={dialogManager.dialogMode}
        selectedDomain={dialogManager.selectedDomain}
        domainToDelete={dialogManager.domainToDelete}
        onFormSubmit={dialogManager.handleFormSubmit}
        onFormCancel={dialogManager.handleFormCancel}
        onConfirmDelete={dialogManager.handleConfirmDelete}
        onCancelDelete={dialogManager.handleCancelDelete}
      />

      <CapabilityDrawer
        capability={hookData.selectedCapability}
        domain={hookData.selectedDomain}
        l1Name={hookData.selectedL1Name}
        getRealizationsForCapability={hookData.getRealizationsForSelectedCapability}
        onClose={hookData.clearCapabilityDetails}
        onChipClick={hookData.handleApplicationClick}
      />

      <ApplicationDrawer componentId={hookData.selectedComponentId} onClose={hookData.clearSelectedComponent} />
    </PageLoadingStates>
  );
}
