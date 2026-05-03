import { zodResolver } from '@hookform/resolvers/zod';
import { Alert, Button, Group, Modal, Stack, Textarea, TextInput } from '@mantine/core';
import React, { useLayoutEffect, useState } from 'react';
import { type FieldErrors, type UseFormRegister, useForm } from 'react-hook-form';
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

interface FormFieldsProps {
  register: UseFormRegister<CreateVendorFormData>;
  errors: FieldErrors<CreateVendorFormData>;
  isPending: boolean;
}

function VendorFields({ register, errors, isPending }: FormFieldsProps) {
  return (
    <>
      <TextInput
        label="Name"
        placeholder="Enter vendor name (e.g., SAP, Microsoft)"
        {...register('name')}
        required
        withAsterisk
        autoFocus
        disabled={isPending}
        error={errors.name?.message}
        data-testid="vendor-name-input"
      />

      <TextInput
        label="Implementation Partner"
        placeholder="Enter implementation partner (optional)"
        {...register('implementationPartner')}
        disabled={isPending}
        error={errors.implementationPartner?.message}
        data-testid="vendor-partner-input"
      />

      <Textarea
        label="Notes"
        placeholder="Enter notes (optional)"
        {...register('notes')}
        rows={3}
        disabled={isPending}
        error={errors.notes?.message}
        data-testid="vendor-notes-input"
      />
    </>
  );
}

interface FormActionsProps {
  isPending: boolean;
  isValid: boolean;
  onCancel: () => void;
}

function FormActions({ isPending, isValid, onCancel }: FormActionsProps) {
  return (
    <Group justify="flex-end" gap="sm">
      <Button variant="default" onClick={onCancel} disabled={isPending} data-testid="create-vendor-cancel">
        Cancel
      </Button>
      <Button type="submit" loading={isPending} disabled={!isValid} data-testid="create-vendor-submit">
        Create
      </Button>
    </Group>
  );
}

function useVendorFormState(
  isOpen: boolean,
  onClose: () => void,
  onCreated?: (vendor: CreatedVendor) => void | Promise<void>,
) {
  const [backendError, setBackendError] = useState<string | null>(null);
  const createMutation = useCreateVendor();
  const form = useForm<CreateVendorFormData>({
    resolver: zodResolver(createVendorSchema),
    defaultValues: DEFAULT_VALUES,
    mode: 'onChange',
  });

  useLayoutEffect(() => {
    if (!isOpen) return;
    form.reset(DEFAULT_VALUES);
    if (backendError !== null) queueMicrotask(() => setBackendError(null));
  }, [isOpen, form, backendError]);

  const onSubmit = async (data: CreateVendorFormData) => {
    setBackendError(null);
    try {
      const created = await createMutation.mutateAsync({
        name: data.name,
        implementationPartner: data.implementationPartner || undefined,
        notes: data.notes || undefined,
      });
      if (onCreated) await onCreated(created);
      onClose();
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'Failed to create vendor');
    }
  };

  return { form, onSubmit, backendError, isPending: createMutation.isPending };
}

export const CreateVendorDialog: React.FC<CreateVendorDialogProps> = ({ isOpen, onClose, onCreated }) => {
  const { form, onSubmit, backendError, isPending } = useVendorFormState(isOpen, onClose, onCreated);
  const {
    register,
    handleSubmit,
    formState: { errors, isValid },
  } = form;

  return (
    <Modal opened={isOpen} onClose={onClose} title="Create Vendor" centered data-testid="create-vendor-dialog">
      <form onSubmit={handleSubmit(onSubmit)}>
        <Stack gap="md">
          <VendorFields register={register} errors={errors} isPending={isPending} />

          {backendError && (
            <Alert color="red" data-testid="create-vendor-error">
              {backendError}
            </Alert>
          )}

          <FormActions isPending={isPending} isValid={isValid} onCancel={onClose} />
        </Stack>
      </form>
    </Modal>
  );
};
