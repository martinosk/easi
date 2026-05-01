import { zodResolver } from '@hookform/resolvers/zod';
import { Alert, Button, Group, Modal, Stack, Textarea, TextInput } from '@mantine/core';
import React, { useLayoutEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { type CreateComponentFormData, createComponentSchema } from '../../../lib/schemas';
import { canEdit } from '../../../utils/hateoas';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { useAddComponentToView } from '../../views/hooks/useViews';
import { useCreateComponent } from '../hooks/useComponents';

interface CreatedComponent {
  id: string;
  name: string;
}

interface CreateComponentDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onCreated?: (component: CreatedComponent) => void | Promise<void>;
}

const DEFAULT_VALUES: CreateComponentFormData = { name: '', description: '' };

export const CreateComponentDialog: React.FC<CreateComponentDialogProps> = ({ isOpen, onClose, onCreated }) => {
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

  useLayoutEffect(() => {
    if (isOpen) {
      reset(DEFAULT_VALUES);
      if (backendError !== null) queueMicrotask(() => setBackendError(null));
    }
  }, [isOpen, reset, backendError]);

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

      if (onCreated) {
        await onCreated(newComponent);
      } else if (currentView && canEdit(currentView)) {
        const defaultPosition = { x: 400, y: 300 };
        await addComponentToViewMutation.mutateAsync({
          viewId: currentView.id,
          request: {
            componentId: newComponent.id,
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
            <Button variant="default" onClick={handleClose} disabled={isCreating} data-testid="create-component-cancel">
              Cancel
            </Button>
            <Button type="submit" loading={isCreating} disabled={!isValid} data-testid="create-component-submit">
              Create Application
            </Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  );
};
