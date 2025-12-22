import React, { useState } from 'react';
import { Modal, TextInput, Textarea, Button, Group, Stack, Alert } from '@mantine/core';
import { useCreateComponent } from '../hooks/useComponents';
import { useAddComponentToView } from '../../views/hooks/useViews';
import { useCurrentView } from '../../../hooks/useCurrentView';
import type { ComponentId, ViewId } from '../../../api/types';

interface CreateComponentDialogProps {
  isOpen: boolean;
  onClose: () => void;
}

export const CreateComponentDialog: React.FC<CreateComponentDialogProps> = ({
  isOpen,
  onClose,
}) => {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [error, setError] = useState<string | null>(null);

  const { currentView } = useCurrentView();
  const createComponentMutation = useCreateComponent();
  const addComponentToViewMutation = useAddComponentToView();

  const isCreating = createComponentMutation.isPending || addComponentToViewMutation.isPending;

  const handleClose = () => {
    setName('');
    setDescription('');
    setError(null);
    onClose();
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!name.trim()) {
      setError('Application name is required');
      return;
    }

    try {
      const newComponent = await createComponentMutation.mutateAsync({
        name: name.trim(),
        description: description.trim() || undefined,
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
      setError(err instanceof Error ? err.message : 'Failed to create application');
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
      <form onSubmit={handleSubmit}>
        <Stack gap="md">
          <TextInput
            label="Name"
            placeholder="Enter application name"
            value={name}
            onChange={(e) => setName(e.currentTarget.value)}
            required
            withAsterisk
            autoFocus
            disabled={isCreating}
            data-testid="component-name-input"
          />

          <Textarea
            label="Description"
            placeholder="Enter application description (optional)"
            value={description}
            onChange={(e) => setDescription(e.currentTarget.value)}
            rows={3}
            disabled={isCreating}
            data-testid="component-description-input"
          />

          {error && (
            <Alert color="red" data-testid="create-component-error">
              {error}
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
              disabled={!name.trim()}
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
