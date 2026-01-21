import { useState, useRef, useEffect } from 'react';
import toast from 'react-hot-toast';
import type { BusinessDomain } from '../../../api/types';

type DialogMode = 'create' | 'edit' | null;

interface UseDomainDialogManagerProps {
  createDomain: (name: string, description?: string, domainArchitectId?: string) => Promise<BusinessDomain>;
  updateDomain: (domain: BusinessDomain, name: string, description?: string, domainArchitectId?: string) => Promise<BusinessDomain>;
  deleteDomain: (domain: BusinessDomain) => Promise<void>;
  onDomainDeleted?: (deletedDomainId: string) => void;
}

export function useDomainDialogManager({
  createDomain,
  updateDomain,
  deleteDomain,
  onDomainDeleted,
}: UseDomainDialogManagerProps) {
  const [dialogMode, setDialogMode] = useState<DialogMode>(null);
  const [selectedDomain, setSelectedDomain] = useState<BusinessDomain | null>(null);
  const [domainToDelete, setDomainToDelete] = useState<BusinessDomain | null>(null);
  const dialogRef = useRef<HTMLDialogElement>(null);

  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;

    if (dialogMode) {
      dialog.showModal();
    } else {
      dialog.close();
    }
  }, [dialogMode]);

  const handleCreateClick = () => {
    setSelectedDomain(null);
    setDialogMode('create');
  };

  const handleEditClick = (domain: BusinessDomain) => {
    setSelectedDomain(domain);
    setDialogMode('edit');
  };

  const handleDeleteClick = (domain: BusinessDomain) => {
    setDomainToDelete(domain);
  };

  const handleFormSubmit = async (name: string, description: string, domainArchitectId?: string) => {
    if (dialogMode === 'create') {
      await createDomain(name, description, domainArchitectId);
      toast.success('Domain created successfully');
    } else if (dialogMode === 'edit' && selectedDomain) {
      await updateDomain(selectedDomain, name, description, domainArchitectId);
      toast.success('Domain updated successfully');
    }
    setDialogMode(null);
    setSelectedDomain(null);
  };

  const handleFormCancel = () => {
    setDialogMode(null);
    setSelectedDomain(null);
  };

  const handleConfirmDelete = async () => {
    if (domainToDelete) {
      try {
        await deleteDomain(domainToDelete);
        toast.success('Domain deleted successfully');
        const deletedId = domainToDelete.id;
        setDomainToDelete(null);
        onDomainDeleted?.(deletedId);
      } catch (err) {
        toast.error(err instanceof Error ? err.message : 'Failed to delete domain');
      }
    }
  };

  const handleCancelDelete = () => {
    setDomainToDelete(null);
  };

  return {
    dialogMode,
    selectedDomain,
    domainToDelete,
    dialogRef,
    handleCreateClick,
    handleEditClick,
    handleDeleteClick,
    handleFormSubmit,
    handleFormCancel,
    handleConfirmDelete,
    handleCancelDelete,
  };
}
