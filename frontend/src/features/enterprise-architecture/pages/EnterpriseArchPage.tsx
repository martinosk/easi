import { useState, useCallback } from 'react';
import { EnterpriseArchHeader } from '../components/EnterpriseArchHeader';
import { EnterpriseArchContent } from '../components/EnterpriseArchContent';
import { CreateEnterpriseCapabilityModal } from '../components/CreateEnterpriseCapabilityModal';
import { ConfirmationDialog } from '../../../components/shared/ConfirmationDialog';
import { useEnterpriseCapabilities } from '../hooks/useEnterpriseCapabilities';
import type { EnterpriseCapability, CreateEnterpriseCapabilityRequest } from '../types';
import { useUserStore } from '../../../store/userStore';
import './EnterpriseArchPage.css';

export function EnterpriseArchPage() {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedCapability, setSelectedCapability] = useState<EnterpriseCapability | null>(null);
  const [capabilityToDelete, setCapabilityToDelete] = useState<EnterpriseCapability | null>(null);

  const hasPermission = useUserStore((state) => state.hasPermission);
  const canRead = hasPermission('enterprise-arch:read');
  const canWrite = hasPermission('enterprise-arch:write');
  const canDelete = hasPermission('enterprise-arch:delete');

  const { capabilities, isLoading, error, createCapability, deleteCapability } = useEnterpriseCapabilities();

  const handleCreateCapability = useCallback(async (request: CreateEnterpriseCapabilityRequest) => {
    await createCapability(request);
    setIsModalOpen(false);
  }, [createCapability]);

  const handleDeleteClick = useCallback((capability: EnterpriseCapability) => {
    setCapabilityToDelete(capability);
  }, []);

  const handleConfirmDelete = useCallback(async () => {
    if (!capabilityToDelete) return;

    await deleteCapability(capabilityToDelete.id, capabilityToDelete.name);

    if (selectedCapability?.id === capabilityToDelete.id) {
      setSelectedCapability(null);
    }
    setCapabilityToDelete(null);
  }, [capabilityToDelete, selectedCapability, deleteCapability]);

  const handleCancelDelete = useCallback(() => {
    setCapabilityToDelete(null);
  }, []);

  const handleSelectCapability = useCallback((capability: EnterpriseCapability) => {
    setSelectedCapability(capability.id === selectedCapability?.id ? null : capability);
  }, [selectedCapability]);

  const handleOpenModal = useCallback(() => {
    setIsModalOpen(true);
  }, []);

  const handleCloseModal = useCallback(() => {
    setIsModalOpen(false);
  }, []);

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
        />
      )}
    </div>
  );
}
