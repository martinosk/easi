import { zodResolver } from '@hookform/resolvers/zod';
import { Alert, Button, Group, Modal, Stack, Textarea, TextInput } from '@mantine/core';
import React, { useLayoutEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { type CreateComponentFormData, createComponentSchema } from '../../../lib/schemas';
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
  const [isHandoffPending, setHandoffPending] = useState(false);

  const createComponentMutation = useCreateComponent();

  const isCreating = createComponentMutation.isPending || isHandoffPending;

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
      setBackendError(null);
      setHandoffPending(false);
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

      if (onCreated) {
        setHandoffPending(true);
        try {
          await onCreated(newComponent);
        } finally {
          setHandoffPending(false);
        }
      }

      handleClose();
    } catch (err) {
      setHandoffPending(false);
      setBackendError(err instanceof Error ? err.message : 'Failed to create application');
    }
  };

  return (
    <Modal opened={isOpen} onClose={handleClose} title="Create Application" centered>
      <form onSubmit={handleSubmit(onSubmit)} data-testid="create-component-dialog">
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
