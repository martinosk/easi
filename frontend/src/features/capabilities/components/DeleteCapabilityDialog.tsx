import React, { useState, useRef, useEffect } from 'react';
import { useAppStore } from '../../../store/appStore';
import type { Capability } from '../../../api/types';

interface DeleteCapabilityDialogProps {
  isOpen: boolean;
  onClose: () => void;
  capability: Capability | null;
  onConfirm?: () => void;
}

export const DeleteCapabilityDialog: React.FC<DeleteCapabilityDialogProps> = ({
  isOpen,
  onClose,
  capability,
  onConfirm,
}) => {
  const [isDeleting, setIsDeleting] = useState(false);
  const [backendError, setBackendError] = useState<string | null>(null);

  const dialogRef = useRef<HTMLDialogElement>(null);
  const deleteCapability = useAppStore((state) => state.deleteCapability);

  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;

    if (isOpen && capability) {
      dialog.showModal();
    } else {
      dialog.close();
    }
  }, [isOpen, capability]);

  const handleClose = () => {
    setBackendError(null);
    onClose();
  };

  const handleConfirm = async () => {
    if (!capability) return;

    setIsDeleting(true);
    setBackendError(null);

    try {
      await deleteCapability(capability.id);
      onConfirm?.();
      handleClose();
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'Failed to delete capability');
    } finally {
      setIsDeleting(false);
    }
  };

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (!isOpen) return;

      if (e.key === 'Escape') {
        handleClose();
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [isOpen]);

  if (!capability) return null;

  return (
    <dialog
      ref={dialogRef}
      className="dialog"
      onClose={handleClose}
      data-testid="delete-capability-dialog"
    >
      <div className="dialog-content">
        <h2 className="dialog-title">Delete Capability?</h2>

        <p>Are you sure you want to delete</p>
        <p className="dialog-item-name">"{capability.name}"</p>
        <p className="dialog-warning">This action cannot be undone.</p>

        {backendError && (
          <div className="error-message" data-testid="delete-capability-error">
            {backendError}
          </div>
        )}

        <div className="dialog-actions">
          <button
            type="button"
            className="btn btn-secondary"
            onClick={handleClose}
            disabled={isDeleting}
            data-testid="delete-capability-cancel"
          >
            Cancel
          </button>
          <button
            type="button"
            className="btn btn-danger"
            onClick={handleConfirm}
            disabled={isDeleting}
            data-testid="delete-capability-submit"
          >
            {isDeleting ? 'Deleting...' : 'Delete'}
          </button>
        </div>
      </div>
    </dialog>
  );
};
