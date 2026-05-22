import { Modal } from '@mantine/core';
import type { BusinessDomain } from '../../../api/types';
import { ConfirmationDialog } from '../../../components/shared/ConfirmationDialog';
import { DomainForm } from './DomainForm';

type DialogMode = 'create' | 'edit' | null;

interface DomainDialogsProps {
  dialogMode: DialogMode;
  selectedDomain: BusinessDomain | null;
  domainToDelete: BusinessDomain | null;
  onFormSubmit: (name: string, description: string, domainArchitectId?: string) => Promise<void>;
  onFormCancel: () => void;
  onConfirmDelete: () => Promise<void>;
  onCancelDelete: () => void;
}

export function DomainDialogs({
  dialogMode,
  selectedDomain,
  domainToDelete,
  onFormSubmit,
  onFormCancel,
  onConfirmDelete,
  onCancelDelete,
}: DomainDialogsProps) {
  return (
    <>
      <Modal
        opened={dialogMode !== null}
        onClose={onFormCancel}
        title={dialogMode === 'create' ? 'Create Domain' : 'Edit Domain'}
        centered
        data-testid="domain-dialog"
      >
        <DomainForm
          mode={dialogMode || 'create'}
          domain={selectedDomain || undefined}
          onSubmit={onFormSubmit}
          onCancel={onFormCancel}
        />
      </Modal>

      {domainToDelete && (
        <ConfirmationDialog
          title="Delete Domain"
          message={`Are you sure you want to delete "${domainToDelete.name}"?`}
          confirmText="Delete"
          cancelText="Cancel"
          onConfirm={onConfirmDelete}
          onCancel={onCancelDelete}
        />
      )}
    </>
  );
}
