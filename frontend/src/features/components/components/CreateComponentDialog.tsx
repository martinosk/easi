import React, { useEffect, useState } from 'react';
import { Modal, TextInput, Textarea, Button, Group, Stack, Alert } from '@mantine/core';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useCreateComponent } from '../hooks/useComponents';
import { useAddComponentToView } from '../../views/hooks/useViews';
import { useCurrentView } from '../../../hooks/useCurrentView';
import { createComponentSchema, type CreateComponentFormData } from '../../../lib/schemas';
import type { ComponentId, ViewId } from '../../../api/types';

interface CreateComponentDialogProps {
  isOpen: boolean;
  onClose: () => void;
}

const DEFAULT_VALUES: CreateComponentFormData = { name: '', description: '' };

export const CreateComponentDialog: React.FC<CreateComponentDialogProps> = ({
  isOpen,
  onClose,
}) => {
  const [backendError, setBackendError] = useState<string | null>(null);

  const { currentView } = useCurrentView();
  const createComponentMutation = useCreateComponent();
  const addComponentToViewMutation = useAddComponentToView();

  const isCreating = createComponentMutation.isPending || addComponentToViewMutation.isPending;

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isValid },
  } = useForm<CreateComponentFormData>({
    resolver: zodResolver(createComponentSchema),
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

  const onSubmit = async (data: CreateComponentFormData) => {
    setBackendError(null);
    try {
      const newComponent = await createComponentMutation.mutateAsync({
        name: data.name,
        description: data.description || undefined,
      });

      if (currentView) {
        const defaultPosition = { x: 400, y: 300 };
        await addComponentToViewMutation.mutateAsync({
          viewId: currentView.id as ViewId,
          request: {
            componentId: newComponent.id as ComponentId,
            ...defaultPosition,
          },
        });
      }

      handleClose();
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'Failed to create application');
    }
  };

  return (
    <Modal
      opened={isOpen}
      onClose={handleClose}
      title="Create Application"
      centered
      data-testid="create-component-dialog"
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
            disabled={isCreating}
            error={errors.name?.message}
            data-testid="component-name-input"
          />

          <Textarea
            label="Description"
            placeholder="Enter application description (optional)"
            {...register('description')}
            rows={3}
            disabled={isCreating}
            error={errors.description?.message}
            data-testid="component-description-input"
          />

          {backendError && (
            <Alert color="red" data-testid="create-component-error">
              {backendError}
            </Alert>
          )}

          <Group justify="flex-end" gap="sm">
            <Button
              variant="default"
              onClick={handleClose}
              disabled={isCreating}
              data-testid="create-component-cancel"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              loading={isCreating}
              disabled={!isValid}
              data-testid="create-component-submit"
            >
              Create Application
            </Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  );
};
