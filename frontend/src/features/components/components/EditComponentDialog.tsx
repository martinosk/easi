import React, { useEffect, useState } from 'react';
import { Modal, TextInput, Textarea, Button, Group, Stack, Alert } from '@mantine/core';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useUpdateComponent } from '../hooks/useComponents';
import { editComponentSchema, type EditComponentFormData } from '../../../lib/schemas';
import type { Component } from '../../../api/types';

interface EditComponentDialogProps {
  isOpen: boolean;
  onClose: () => void;
  component: Component | null;
}

const DEFAULT_VALUES: EditComponentFormData = { name: '', description: '' };

export const EditComponentDialog: React.FC<EditComponentDialogProps> = ({
  isOpen,
  onClose,
  component,
}) => {
  const [backendError, setBackendError] = useState<string | null>(null);
  const updateComponentMutation = useUpdateComponent();

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isValid },
  } = useForm<EditComponentFormData>({
    resolver: zodResolver(editComponentSchema),
    defaultValues: DEFAULT_VALUES,
    mode: 'onChange',
  });

  useEffect(() => {
    if (isOpen && component) {
      reset({
        name: component.name,
        description: component.description || '',
      });
      setBackendError(null);
    }
  }, [isOpen, component, reset]);

  const handleClose = () => {
    onClose();
  };

  const onSubmit = async (data: EditComponentFormData) => {
    if (!component) {
      setBackendError('No application selected');
      return;
    }
    setBackendError(null);
    try {
      await updateComponentMutation.mutateAsync({
        component,
        request: {
          name: data.name,
          description: data.description || undefined,
        },
      });
      handleClose();
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'Failed to update application');
    }
  };

  return (
    <Modal
      opened={isOpen}
      onClose={handleClose}
      title="Edit Application"
      centered
    >
      <form onSubmit={handleSubmit(onSubmit)}>
        <Stack gap="md">
          <TextInput
            label="Name"
            placeholder="Enter application name"
            {...register('name')}
            required
            withAsterisk
            autoFocus
            disabled={updateComponentMutation.isPending}
            error={errors.name?.message}
          />

          <Textarea
            label="Description"
            placeholder="Enter application description (optional)"
            {...register('description')}
            rows={3}
            disabled={updateComponentMutation.isPending}
            error={errors.description?.message}
          />

          {backendError && (
            <Alert color="red">
              {backendError}
            </Alert>
          )}

          <Group justify="flex-end" gap="sm">
            <Button
              variant="default"
              onClick={handleClose}
              disabled={updateComponentMutation.isPending}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              loading={updateComponentMutation.isPending}
              disabled={!isValid}
            >
              Save Changes
            </Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  );
};
