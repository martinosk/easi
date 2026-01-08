import React, { useEffect, useState } from 'react';
import { Modal, TextInput, Button, Group, Stack, Alert } from '@mantine/core';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useAddCapabilityTag } from '../hooks/useCapabilities';
import { addTagSchema, type AddTagFormData } from '../../../lib/schemas';
import type { CapabilityId } from '../../../api/types';

interface AddTagDialogProps {
  isOpen: boolean;
  onClose: () => void;
  capabilityId: string;
}

const DEFAULT_VALUES: AddTagFormData = { tag: '' };

export const AddTagDialog: React.FC<AddTagDialogProps> = ({
  isOpen,
  onClose,
  capabilityId,
}) => {
  const [backendError, setBackendError] = useState<string | null>(null);
  const addTagMutation = useAddCapabilityTag();

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isValid },
  } = useForm<AddTagFormData>({
    resolver: zodResolver(addTagSchema),
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

  const onSubmit = async (data: AddTagFormData) => {
    setBackendError(null);
    try {
      await addTagMutation.mutateAsync({
        id: capabilityId as CapabilityId,
        request: { tag: data.tag },
      });
      handleClose();
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'Failed to add tag');
    }
  };

  return (
    <Modal
      opened={isOpen}
      onClose={handleClose}
      title="Add Tag"
      centered
      data-testid="add-tag-dialog"
    >
      <form onSubmit={handleSubmit(onSubmit)}>
        <Stack gap="md">
          <TextInput
            label="Tag Name"
            placeholder="Enter tag name"
            {...register('tag')}
            required
            withAsterisk
            autoFocus
            disabled={addTagMutation.isPending}
            error={errors.tag?.message}
            data-testid="tag-name-input"
          />

          {backendError && (
            <Alert color="red" data-testid="add-tag-error">
              {backendError}
            </Alert>
          )}

          <Group justify="flex-end" gap="sm">
            <Button
              variant="default"
              onClick={handleClose}
              disabled={addTagMutation.isPending}
              data-testid="add-tag-cancel"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              loading={addTagMutation.isPending}
              disabled={!isValid}
              data-testid="add-tag-submit"
            >
              Add
            </Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  );
};
