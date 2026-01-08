import React, { useEffect, useState } from 'react';
import { Modal, TextInput, Button, Group, Stack, Alert } from '@mantine/core';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useAddCapabilityExpert } from '../hooks/useCapabilities';
import { addExpertSchema, type AddExpertFormData } from '../../../lib/schemas';
import type { CapabilityId } from '../../../api/types';

interface AddExpertDialogProps {
  isOpen: boolean;
  onClose: () => void;
  capabilityId: string;
}

const DEFAULT_VALUES: AddExpertFormData = { name: '', role: '', contact: '' };

export const AddExpertDialog: React.FC<AddExpertDialogProps> = ({
  isOpen,
  onClose,
  capabilityId,
}) => {
  const [backendError, setBackendError] = useState<string | null>(null);
  const addExpertMutation = useAddCapabilityExpert();

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isValid },
  } = useForm<AddExpertFormData>({
    resolver: zodResolver(addExpertSchema),
    defaultValues: DEFAULT_VALUES,
    mode: 'onChange',
  });

  useEffect(() => {
    if (isOpen) {
      reset(DEFAULT_VALUES);
      setBackendError(null);
    }
  }, [isOpen, reset]);

  const handleClose = () => {
    onClose();
  };

  const onSubmit = async (data: AddExpertFormData) => {
    setBackendError(null);
    try {
      await addExpertMutation.mutateAsync({
        id: capabilityId as CapabilityId,
        request: {
          expertName: data.name,
          expertRole: data.role,
          contactInfo: data.contact,
        },
      });
      handleClose();
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'Failed to add expert');
    }
  };

  return (
    <Modal
      opened={isOpen}
      onClose={handleClose}
      title="Add Expert"
      centered
      data-testid="add-expert-dialog"
    >
      <form onSubmit={handleSubmit(onSubmit)}>
        <Stack gap="md">
          <TextInput
            label="Name"
            placeholder="Enter expert name"
            {...register('name')}
            required
            withAsterisk
            autoFocus
            disabled={addExpertMutation.isPending}
            error={errors.name?.message}
            data-testid="expert-name-input"
          />

          <TextInput
            label="Role"
            placeholder="Enter expert role"
            {...register('role')}
            required
            withAsterisk
            disabled={addExpertMutation.isPending}
            error={errors.role?.message}
            data-testid="expert-role-input"
          />

          <TextInput
            label="Contact"
            placeholder="Enter contact information"
            {...register('contact')}
            required
            withAsterisk
            disabled={addExpertMutation.isPending}
            error={errors.contact?.message}
            data-testid="expert-contact-input"
          />

          {backendError && (
            <Alert color="red" data-testid="add-expert-error">
              {backendError}
            </Alert>
          )}

          <Group justify="flex-end" gap="sm">
            <Button
              variant="default"
              onClick={handleClose}
              disabled={addExpertMutation.isPending}
              data-testid="add-expert-cancel"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              loading={addExpertMutation.isPending}
              disabled={!isValid}
              data-testid="add-expert-submit"
            >
              Add
            </Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  );
};
