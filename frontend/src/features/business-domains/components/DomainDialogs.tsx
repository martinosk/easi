import type { RefObject } from 'react';
import type { BusinessDomain } from '../../../api/types';
import { DomainForm } from './DomainForm';
import { ConfirmationDialog } from '../../../components/shared/ConfirmationDialog';

type DialogMode = 'create' | 'edit' | null;

interface DomainDialogsProps {
  dialogMode: DialogMode;
  selectedDomain: BusinessDomain | null;
  domainToDelete: BusinessDomain | null;
  dialogRef: RefObject<HTMLDialogElement | null>;
  onFormSubmit: (name: string, description: string, domainArchitectId?: string) => Promise<void>;
  onFormCancel: () => void;
  onConfirmDelete: () => Promise<void>;
  onCancelDelete: () => void;
}

export function DomainDialogs({
  dialogMode,
  selectedDomain,
  domainToDelete,
  dialogRef,
  onFormSubmit,
  onFormCancel,
  onConfirmDelete,
  onCancelDelete,
}: DomainDialogsProps) {
  return (
    <>
      <dialog ref={dialogRef} className="dialog" onClose={onFormCancel} data-testid="domain-dialog">
        <div className="dialog-content">
          <h2 className="dialog-title">{dialogMode === 'create' ? 'Create Domain' : 'Edit Domain'}</h2>
          <DomainForm
            mode={dialogMode || 'create'}
            domain={selectedDomain || undefined}
            onSubmit={onFormSubmit}
            onCancel={onFormCancel}
          />
        </div>
      </dialog>

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
