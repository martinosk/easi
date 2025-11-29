import { useState } from 'react';
import toast from 'react-hot-toast';
import { DomainList } from '../components/DomainList';
import { DomainForm } from '../components/DomainForm';
import { ConfirmationDialog } from '../../../components/shared/ConfirmationDialog';
import { useBusinessDomains } from '../hooks/useBusinessDomains';
import type { BusinessDomain } from '../../../api/types';

type DialogMode = 'create' | 'edit' | null;

export function BusinessDomainsPage() {
  const { domains, isLoading, error, createDomain, updateDomain, deleteDomain } = useBusinessDomains();
  const [dialogMode, setDialogMode] = useState<DialogMode>(null);
  const [selectedDomain, setSelectedDomain] = useState<BusinessDomain | null>(null);
  const [domainToDelete, setDomainToDelete] = useState<BusinessDomain | null>(null);

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

  const handleViewClick = (domain: BusinessDomain) => {
    window.location.hash = `#/business-domains/${domain.id}`;
  };

  const handleFormSubmit = async (name: string, description: string) => {
    if (dialogMode === 'create') {
      await createDomain(name, description);
      toast.success('Domain created successfully');
    } else if (dialogMode === 'edit' && selectedDomain) {
      await updateDomain(selectedDomain, name, description);
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
        setDomainToDelete(null);
      } catch (err) {
        toast.error(err instanceof Error ? err.message : 'Failed to delete domain');
      }
    }
  };

  const handleCancelDelete = () => {
    setDomainToDelete(null);
  };

  if (isLoading && domains.length === 0) {
    return (
      <div className="page-container">
        <div className="loading-message">Loading business domains...</div>
      </div>
    );
  }

  if (error && domains.length === 0) {
    return (
      <div className="page-container">
        <div className="error-message" data-testid="domains-error">
          {error.message}
        </div>
      </div>
    );
  }

  return (
    <div className="page-container" data-testid="business-domains-page">
      <div className="page-header">
        <h1>Business Domains</h1>
        <div className="flex gap-2">
          <a
            href="#/business-domains/visualization"
            className="btn btn-secondary"
            data-testid="visualization-link"
          >
            <svg className="w-5 h-5 inline mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 5a1 1 0 011-1h4a1 1 0 011 1v7a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM14 5a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1V5zM4 16a1 1 0 011-1h4a1 1 0 011 1v3a1 1 0 01-1 1H5a1 1 0 01-1-1v-3zM14 13a1 1 0 011-1h4a1 1 0 011 1v6a1 1 0 01-1 1h-4a1 1 0 01-1-1v-6z" />
            </svg>
            Visualization
          </a>
          <button
            type="button"
            className="btn btn-primary"
            onClick={handleCreateClick}
            data-testid="create-domain-button"
          >
            Create Domain
          </button>
        </div>
      </div>

      <DomainList
        domains={domains}
        onEdit={handleEditClick}
        onDelete={handleDeleteClick}
        onView={handleViewClick}
      />

      {dialogMode && (
        <dialog open className="dialog" data-testid="domain-dialog">
          <div className="dialog-content">
            <h2 className="dialog-title">{dialogMode === 'create' ? 'Create Domain' : 'Edit Domain'}</h2>
            <DomainForm
              mode={dialogMode}
              domain={selectedDomain || undefined}
              onSubmit={handleFormSubmit}
              onCancel={handleFormCancel}
            />
          </div>
        </dialog>
      )}

      {domainToDelete && (
        <ConfirmationDialog
          title="Delete Domain"
          message={`Are you sure you want to delete "${domainToDelete.name}"?`}
          confirmText="Delete"
          cancelText="Cancel"
          onConfirm={handleConfirmDelete}
          onCancel={handleCancelDelete}
        />
      )}
    </div>
  );
}
