import React, { useState, useEffect } from 'react';
import { Modal, TextInput, Textarea, Button, Group, Stack, Alert } from '@mantine/core';
import { useAppStore } from '../../../store/appStore';
import type { Component } from '../../../api/types';

interface EditComponentDialogProps {
  isOpen: boolean;
  onClose: () => void;
  component: Component | null;
}

export const EditComponentDialog: React.FC<EditComponentDialogProps> = ({
  isOpen,
  onClose,
  component,
}) => {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [isUpdating, setIsUpdating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const updateComponent = useAppStore((state) => state.updateComponent);

  useEffect(() => {
    if (component) {
      setName(component.name);
      setDescription(component.description || '');
    }
  }, [component]);

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

    if (!component) {
      setError('No application selected');
      return;
    }

    setIsUpdating(true);

    try {
      await updateComponent(component.id, {
        name: name.trim(),
        description: description.trim() || undefined,
      });
      handleClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update application');
    } finally {
      setIsUpdating(false);
    }
  };

  return (
    <Modal
      opened={isOpen}
      onClose={handleClose}
      title="Edit Application"
      centered
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
            disabled={isUpdating}
          />

          <Textarea
            label="Description"
            placeholder="Enter application description (optional)"
            value={description}
            onChange={(e) => setDescription(e.currentTarget.value)}
            rows={3}
            disabled={isUpdating}
          />

          {error && (
            <Alert color="red">
              {error}
            </Alert>
          )}

          <Group justify="flex-end" gap="sm">
            <Button
              variant="default"
              onClick={handleClose}
              disabled={isUpdating}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              loading={isUpdating}
              disabled={!name.trim()}
            >
              Save Changes
            </Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  );
};
