import React, { useState } from 'react';
import { Modal, Text, Button, Group, Stack, Alert } from '@mantine/core';
import { useAppStore } from '../../../store/appStore';
import type { Capability } from '../../../api/types';

interface DeleteCapabilityDialogProps {
  isOpen: boolean;
  onClose: () => void;
  capability: Capability | null;
  onConfirm?: () => void;
  capabilitiesToDelete?: Capability[];
}

export const DeleteCapabilityDialog: React.FC<DeleteCapabilityDialogProps> = ({
  isOpen,
  onClose,
  capability,
  onConfirm,
  capabilitiesToDelete = [],
}) => {
  const [isDeleting, setIsDeleting] = useState(false);
  const [backendError, setBackendError] = useState<string | null>(null);

  const deleteCapability = useAppStore((state) => state.deleteCapability);

  const handleClose = () => {
    setBackendError(null);
    onClose();
  };

  const handleConfirm = async () => {
    if (!capability) return;

    setIsDeleting(true);
    setBackendError(null);

    try {
      const capsToDelete = capabilitiesToDelete.length > 0 ? capabilitiesToDelete : [capability];
      for (const cap of capsToDelete) {
        await deleteCapability(cap.id);
      }
      onConfirm?.();
      handleClose();
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'Failed to delete capability');
    } finally {
      setIsDeleting(false);
    }
  };

  if (!capability) return null;

  const isMultiDelete = capabilitiesToDelete.length > 1;
  const deleteCount = capabilitiesToDelete.length > 0 ? capabilitiesToDelete.length : 1;

  return (
    <Modal
      opened={isOpen}
      onClose={handleClose}
      title="Delete Capability?"
      centered
      data-testid="delete-capability-dialog"
    >
      <Stack gap="md">
        {isMultiDelete ? (
          <Text>Are you sure you want to delete {deleteCount} capabilities?</Text>
        ) : (
          <>
            <Text>Are you sure you want to delete</Text>
            <Text fw={600} size="lg">"{capability.name}"</Text>
          </>
        )}
        <Text c="orange" size="sm">This action cannot be undone.</Text>

        {backendError && (
          <Alert color="red" data-testid="delete-capability-error">
            {backendError}
          </Alert>
        )}

        <Group justify="flex-end" gap="sm">
          <Button
            variant="default"
            onClick={handleClose}
            disabled={isDeleting}
            data-testid="delete-capability-cancel"
          >
            Cancel
          </Button>
          <Button
            color="red"
            onClick={handleConfirm}
            loading={isDeleting}
            data-testid="delete-capability-submit"
          >
            Delete
          </Button>
        </Group>
      </Stack>
    </Modal>
  );
};
