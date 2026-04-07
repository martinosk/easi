import { zodResolver } from '@hookform/resolvers/zod';
import { Alert, Button, Group, Modal, Stack, Textarea, TextInput } from '@mantine/core';
import React, { useLayoutEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import type { Vendor, VendorId } from '../../../api/types';
import { type EditVendorFormData, editVendorSchema } from '../../../lib/schemas';
import { useUpdateVendor } from '../hooks/useVendors';

interface EditVendorDialogProps {
  isOpen: boolean;
  onClose: () => void;
  vendor: Vendor | null;
}

export const EditVendorDialog: React.FC<EditVendorDialogProps> = ({ isOpen, onClose, vendor }) => {
  const [backendError, setBackendError] = useState<string | null>(null);
  const updateMutation = useUpdateVendor();

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isValid },
  } = useForm<EditVendorFormData>({
    resolver: zodResolver(editVendorSchema),
    mode: 'onChange',
  });

  useLayoutEffect(() => {
    if (isOpen && vendor) {
      reset({
        name: vendor.name,
        implementationPartner: vendor.implementationPartner || '',
        notes: vendor.notes || '',
      });
      if (backendError !== null) queueMicrotask(() => setBackendError(null));
    }
  }, [isOpen, vendor, reset, backendError]);

  const handleClose = () => {
    onClose();
  };

  const onSubmit = async (data: EditVendorFormData) => {
    if (!vendor) return;

    setBackendError(null);
    try {
      await updateMutation.mutateAsync({
        id: vendor.id as VendorId,
        request: {
          name: data.name,
          implementationPartner: data.implementationPartner || undefined,
          notes: data.notes || undefined,
        },
      });
      handleClose();
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'Failed to update vendor');
    }
  };

  if (!vendor) return null;

  return (
    <Modal opened={isOpen} onClose={handleClose} title="Edit Vendor" centered data-testid="edit-vendor-dialog">
      <form onSubmit={handleSubmit(onSubmit)}>
        <Stack gap="md">
          <TextInput
            label="Name"
            placeholder="Enter vendor name"
            {...register('name')}
            required
            withAsterisk
            autoFocus
            disabled={updateMutation.isPending}
            error={errors.name?.message}
            data-testid="edit-vendor-name-input"
          />

          <TextInput
            label="Implementation Partner"
            placeholder="Enter implementation partner (optional)"
            {...register('implementationPartner')}
            disabled={updateMutation.isPending}
            error={errors.implementationPartner?.message}
            data-testid="edit-vendor-partner-input"
          />

          <Textarea
            label="Notes"
            placeholder="Enter notes (optional)"
            {...register('notes')}
            rows={3}
            disabled={updateMutation.isPending}
            error={errors.notes?.message}
            data-testid="edit-vendor-notes-input"
          />

          {backendError && (
            <Alert color="red" data-testid="edit-vendor-error">
              {backendError}
            </Alert>
          )}

          <Group justify="flex-end" gap="sm">
            <Button
              variant="default"
              onClick={handleClose}
              disabled={updateMutation.isPending}
              data-testid="edit-vendor-cancel"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              loading={updateMutation.isPending}
              disabled={!isValid}
              data-testid="edit-vendor-submit"
            >
              Save Changes
            </Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  );
};
