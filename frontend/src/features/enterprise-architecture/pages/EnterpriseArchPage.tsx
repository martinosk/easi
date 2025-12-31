import { useState, useCallback, useMemo } from 'react';
import { EnterpriseArchHeader } from '../components/EnterpriseArchHeader';
import { EnterpriseArchContent } from '../components/EnterpriseArchContent';
import { CreateEnterpriseCapabilityModal } from '../components/CreateEnterpriseCapabilityModal';
import { ConfirmationDialog } from '../../../components/shared/ConfirmationDialog';
import { useEnterpriseCapabilities } from '../hooks/useEnterpriseCapabilities';
import { useDomainCapabilityLinking } from '../hooks/useDomainCapabilityLinking';
import { getErrorMessage } from '../utils/errorMessages';
import type { EnterpriseCapability, CreateEnterpriseCapabilityRequest, EnterpriseCapabilityId } from '../types';
import type { Capability } from '../../../api/types';
import { useUserStore } from '../../../store/userStore';
import './EnterpriseArchPage.css';

function useEnterpriseArchPermissions() {
  const hasPermission = useUserStore((state) => state.hasPermission);
  return useMemo(() => ({
    canRead: hasPermission('enterprise-arch:read'),
    canWrite: hasPermission('enterprise-arch:write'),
    canDelete: hasPermission('enterprise-arch:delete'),
  }), [hasPermission]);
}

export function EnterpriseArchPage() {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedCapabilityId, setSelectedCapabilityId] = useState<EnterpriseCapabilityId | null>(null);
  const [capabilityToDelete, setCapabilityToDelete] = useState<EnterpriseCapability | null>(null);
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const [isDockPanelOpen, setIsDockPanelOpen] = useState(false);

  const { canRead, canWrite, canDelete } = useEnterpriseArchPermissions();

  const { capabilities, isLoading, error, createCapability, deleteCapability } = useEnterpriseCapabilities();

  const selectedCapability = useMemo(
    () => capabilities.find((c) => c.id === selectedCapabilityId) || null,
    [capabilities, selectedCapabilityId]
  );

  const {
    domainCapabilities,
    linkStatuses,
    isLoading: isLoadingDomainCapabilities,
    linkCapability,
  } = useDomainCapabilityLinking(isDockPanelOpen);

  const handleCreateCapability = useCallback(async (request: CreateEnterpriseCapabilityRequest) => {
    await createCapability(request);
    setIsModalOpen(false);
  }, [createCapability]);

  const handleDeleteClick = useCallback((capability: EnterpriseCapability) => {
    setCapabilityToDelete(capability);
    setDeleteError(null);
  }, []);

  const handleConfirmDelete = useCallback(async () => {
    if (!capabilityToDelete) return;

    try {
      setDeleteError(null);
      await deleteCapability(capabilityToDelete.id, capabilityToDelete.name);
      if (selectedCapabilityId === capabilityToDelete.id) {
        setSelectedCapabilityId(null);
      }
      setCapabilityToDelete(null);
    } catch (err) {
      setDeleteError(getErrorMessage(err, 'Failed to delete capability'));
    }
  }, [capabilityToDelete, selectedCapabilityId, deleteCapability]);

  const handleCancelDelete = useCallback(() => {
    setCapabilityToDelete(null);
    setDeleteError(null);
  }, []);

  const handleSelectCapability = useCallback((capability: EnterpriseCapability) => {
    setSelectedCapabilityId(capability.id === selectedCapabilityId ? null : capability.id);
  }, [selectedCapabilityId]);

  const handleOpenModal = useCallback(() => {
    setIsModalOpen(true);
  }, []);

  const handleCloseModal = useCallback(() => {
    setIsModalOpen(false);
  }, []);

  const handleToggleDockPanel = useCallback(() => {
    setIsDockPanelOpen((prev) => !prev);
  }, []);

  const handleCloseDockPanel = useCallback(() => {
    setIsDockPanelOpen(false);
  }, []);

  const handleLinkCapability = useCallback(
    async (enterpriseCapabilityId: EnterpriseCapabilityId, domainCapability: Capability) => {
      await linkCapability(enterpriseCapabilityId, domainCapability);
    },
    [linkCapability]
  );

  if (!canRead) {
    return (
      <div className="enterprise-arch-page">
        <div className="enterprise-arch-container">
          <div className="error-message">You do not have permission to view enterprise architecture.</div>
        </div>
      </div>
    );
  }

  return (
    <div className="enterprise-arch-page">
      <div className="enterprise-arch-container">
        <EnterpriseArchHeader
          canWrite={canWrite}
          onCreateNew={handleOpenModal}
          isDockPanelOpen={isDockPanelOpen}
          onToggleDockPanel={handleToggleDockPanel}
        />
        <EnterpriseArchContent
          isLoading={isLoading}
          error={error?.message || null}
          capabilities={capabilities}
          selectedCapability={selectedCapability}
          canWrite={canWrite}
          canDelete={canDelete}
          onSelect={handleSelectCapability}
          onDelete={handleDeleteClick}
          onCreateNew={handleOpenModal}
          isDockPanelOpen={isDockPanelOpen}
          domainCapabilities={domainCapabilities}
          linkStatuses={linkStatuses}
          isLoadingDomainCapabilities={isLoadingDomainCapabilities}
          onCloseDockPanel={handleCloseDockPanel}
          onLinkCapability={handleLinkCapability}
        />
      </div>
      {canWrite && (
        <CreateEnterpriseCapabilityModal
          isOpen={isModalOpen}
          onClose={handleCloseModal}
          onSubmit={handleCreateCapability}
        />
      )}
      {capabilityToDelete && (
        <ConfirmationDialog
          title="Delete Enterprise Capability"
          message={`Are you sure you want to delete "${capabilityToDelete.name}"?`}
          confirmText="Delete"
          cancelText="Cancel"
          onConfirm={handleConfirmDelete}
          onCancel={handleCancelDelete}
          error={deleteError}
        />
      )}
    </div>
  );
}
