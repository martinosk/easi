import React, { useEffect, useState } from 'react';
import { Modal, TextInput, Textarea, Button, Group, Stack, Alert } from '@mantine/core';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useUpdateVendor } from '../hooks/useVendors';
import { editVendorSchema, type EditVendorFormData } from '../../../lib/schemas';
import type { Vendor, VendorId } from '../../../api/types';

interface EditVendorDialogProps {
  isOpen: boolean;
  onClose: () => void;
  vendor: Vendor | null;
}

export const EditVendorDialog: React.FC<EditVendorDialogProps> = ({
  isOpen,
  onClose,
  vendor,
}) => {
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

  useEffect(() => {
    if (isOpen && vendor) {
      reset({
        name: vendor.name,
        implementationPartner: vendor.implementationPartner || '',
        notes: vendor.notes || '',
      });
      setBackendError(null);
    }
  }, [isOpen, vendor, reset]);

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
    <Modal
      opened={isOpen}
      onClose={handleClose}
      title="Edit Vendor"
      centered
      data-testid="edit-vendor-dialog"
    >
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
