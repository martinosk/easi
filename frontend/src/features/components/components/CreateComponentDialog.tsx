import React, { useState } from 'react';
import { Modal, TextInput, Textarea, Button, Group, Stack, Alert } from '@mantine/core';
import { useAppStore } from '../../../store/appStore';

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
  const [isCreating, setIsCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const createComponent = useAppStore((state) => state.createComponent);

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

    setIsCreating(true);

    try {
      await createComponent({
        name: name.trim(),
        description: description.trim() || undefined,
      });
      handleClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create application');
    } finally {
      setIsCreating(false);
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
