import React, { useState, useEffect } from 'react';
import { Modal, TextInput, Textarea, Button, Group, Stack, Alert } from '@mantine/core';
import { useUpdateRelation } from '../hooks/useRelations';
import type { Relation, RelationId } from '../../../api/types';

interface EditRelationDialogProps {
  isOpen: boolean;
  onClose: () => void;
  relation: Relation | null;
}

export const EditRelationDialog: React.FC<EditRelationDialogProps> = ({
  isOpen,
  onClose,
  relation,
}) => {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [error, setError] = useState<string | null>(null);

  const updateRelationMutation = useUpdateRelation();
  const isUpdating = updateRelationMutation.isPending;

  useEffect(() => {
    if (relation) {
      setName(relation.name || '');
      setDescription(relation.description || '');
    }
  }, [relation]);

  const handleClose = () => {
    setName('');
    setDescription('');
    setError(null);
    onClose();
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!relation) {
      setError('No relation selected');
      return;
    }

    try {
      await updateRelationMutation.mutateAsync({
        id: relation.id as RelationId,
        request: {
          name: name.trim() || undefined,
          description: description.trim() || undefined,
        },
      });
      handleClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update relation');
    }
  };

  return (
    <Modal
      opened={isOpen}
      onClose={handleClose}
      title="Edit Relation"
      centered
    >
      <form onSubmit={handleSubmit}>
        <Stack gap="md">
          <TextInput
            label="Name"
            placeholder="Enter relation name (optional)"
            value={name}
            onChange={(e) => setName(e.currentTarget.value)}
            disabled={isUpdating}
            autoFocus
          />

          <Textarea
            label="Description"
            placeholder="Enter relation description (optional)"
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
            >
              Save Changes
            </Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  );
};
