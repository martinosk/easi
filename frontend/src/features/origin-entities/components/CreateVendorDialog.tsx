import { zodResolver } from '@hookform/resolvers/zod';
import { Alert, Button, Group, Modal, Stack, Textarea, TextInput } from '@mantine/core';
import React, { useLayoutEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { type CreateVendorFormData, createVendorSchema } from '../../../lib/schemas';
import { useCreateVendor } from '../hooks/useVendors';

interface CreatedVendor {
  id: string;
  name: string;
}

interface CreateVendorDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onCreated?: (vendor: CreatedVendor) => void | Promise<void>;
}

const DEFAULT_VALUES: CreateVendorFormData = {
  name: '',
  implementationPartner: '',
  notes: '',
};

export const CreateVendorDialog: React.FC<CreateVendorDialogProps> = ({ isOpen, onClose, onCreated }) => {
  const [backendError, setBackendError] = useState<string | null>(null);
  const createMutation = useCreateVendor();

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isValid },
  } = useForm<CreateVendorFormData>({
    resolver: zodResolver(createVendorSchema),
    defaultValues: DEFAULT_VALUES,
    mode: 'onChange',
  });

  useLayoutEffect(() => {
    if (isOpen) {
      reset(DEFAULT_VALUES);
      if (backendError !== null) queueMicrotask(() => setBackendError(null));
    }
  }, [isOpen, reset, backendError]);

  const handleClose = () => {
    onClose();
  };

  const onSubmit = async (data: CreateVendorFormData) => {
    setBackendError(null);
    try {
      const created = await createMutation.mutateAsync({
        name: data.name,
        implementationPartner: data.implementationPartner || undefined,
        notes: data.notes || undefined,
      });
      if (onCreated) await onCreated(created);
      handleClose();
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'Failed to create vendor');
    }
  };

  return (
    <Modal opened={isOpen} onClose={handleClose} title="Create Vendor" centered data-testid="create-vendor-dialog">
      <form onSubmit={handleSubmit(onSubmit)}>
        <Stack gap="md">
          <TextInput
            label="Name"
            placeholder="Enter vendor name (e.g., SAP, Microsoft)"
            {...register('name')}
            required
            withAsterisk
            autoFocus
            disabled={createMutation.isPending}
            error={errors.name?.message}
            data-testid="vendor-name-input"
          />

          <TextInput
            label="Implementation Partner"
            placeholder="Enter implementation partner (optional)"
            {...register('implementationPartner')}
            disabled={createMutation.isPending}
            error={errors.implementationPartner?.message}
            data-testid="vendor-partner-input"
          />

          <Textarea
            label="Notes"
            placeholder="Enter notes (optional)"
            {...register('notes')}
            rows={3}
            disabled={createMutation.isPending}
            error={errors.notes?.message}
            data-testid="vendor-notes-input"
          />

          {backendError && (
            <Alert color="red" data-testid="create-vendor-error">
              {backendError}
            </Alert>
          )}

          <Group justify="flex-end" gap="sm">
            <Button
              variant="default"
              onClick={handleClose}
              disabled={createMutation.isPending}
              data-testid="create-vendor-cancel"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              loading={createMutation.isPending}
              disabled={!isValid}
              data-testid="create-vendor-submit"
            >
              Create
            </Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  );
};
