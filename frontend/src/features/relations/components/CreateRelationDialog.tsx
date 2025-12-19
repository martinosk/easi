import React, { useState, useEffect } from 'react';
import { Modal, Select, TextInput, Textarea, Button, Group, Stack, Alert } from '@mantine/core';
import { useAppStore } from '../../../store/appStore';

interface CreateRelationDialogProps {
  isOpen: boolean;
  onClose: () => void;
  sourceComponentId?: string;
  targetComponentId?: string;
}

export const CreateRelationDialog: React.FC<CreateRelationDialogProps> = ({
  isOpen,
  onClose,
  sourceComponentId: initialSource,
  targetComponentId: initialTarget,
}) => {
  const [sourceComponentId, setSourceComponentId] = useState(initialSource || '');
  const [targetComponentId, setTargetComponentId] = useState(initialTarget || '');
  const [relationType, setRelationType] = useState<'Triggers' | 'Serves'>('Triggers');
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [isCreating, setIsCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const components = useAppStore((state) => state.components);
  const createRelation = useAppStore((state) => state.createRelation);

  useEffect(() => {
    if (initialSource) setSourceComponentId(initialSource);
    if (initialTarget) setTargetComponentId(initialTarget);
  }, [initialSource, initialTarget]);

  const handleClose = () => {
    setSourceComponentId('');
    setTargetComponentId('');
    setRelationType('Triggers');
    setName('');
    setDescription('');
    setError(null);
    onClose();
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!sourceComponentId || !targetComponentId) {
      setError('Both source and target components are required');
      return;
    }

    if (sourceComponentId === targetComponentId) {
      setError('Source and target components must be different');
      return;
    }

    setIsCreating(true);

    try {
      await createRelation({
        sourceComponentId: sourceComponentId as import('../../../api/types').ComponentId,
        targetComponentId: targetComponentId as import('../../../api/types').ComponentId,
        relationType,
        name: name.trim() || undefined,
        description: description.trim() || undefined,
      });
      handleClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create relation');
    } finally {
      setIsCreating(false);
    }
  };

  const componentOptions = components.map((component) => ({
    value: component.id,
    label: component.name,
  }));

  return (
    <Modal
      opened={isOpen}
      onClose={handleClose}
      title="Create Relation"
      centered
      data-testid="create-relation-dialog"
    >
      <form onSubmit={handleSubmit}>
        <Stack gap="md">
          <Select
            label="Source Component"
            placeholder="Select source component"
            value={sourceComponentId}
            onChange={(value) => setSourceComponentId(value || '')}
            data={componentOptions}
            required
            withAsterisk
            disabled={isCreating || !!initialSource}
            data-testid="relation-source-select"
            searchable
          />

          <Select
            label="Target Component"
            placeholder="Select target component"
            value={targetComponentId}
            onChange={(value) => setTargetComponentId(value || '')}
            data={componentOptions}
            required
            withAsterisk
            disabled={isCreating || !!initialTarget}
            data-testid="relation-target-select"
            searchable
          />

          <Select
            label="Relation Type"
            value={relationType}
            onChange={(value) => setRelationType(value as 'Triggers' | 'Serves')}
            data={[
              { value: 'Triggers', label: 'Triggers' },
              { value: 'Serves', label: 'Serves' },
            ]}
            required
            withAsterisk
            disabled={isCreating}
            data-testid="relation-type-select"
          />

          <TextInput
            label="Name"
            placeholder="Enter relation name (optional)"
            value={name}
            onChange={(e) => setName(e.currentTarget.value)}
            disabled={isCreating}
            data-testid="relation-name-input"
          />

          <Textarea
            label="Description"
            placeholder="Enter relation description (optional)"
            value={description}
            onChange={(e) => setDescription(e.currentTarget.value)}
            rows={3}
            disabled={isCreating}
            data-testid="relation-description-input"
          />

          {error && (
            <Alert color="red" data-testid="create-relation-error">
              {error}
            </Alert>
          )}

          <Group justify="flex-end" gap="sm">
            <Button
              variant="default"
              onClick={handleClose}
              disabled={isCreating}
              data-testid="create-relation-cancel"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              loading={isCreating}
              disabled={!sourceComponentId || !targetComponentId}
              data-testid="create-relation-submit"
            >
              Create Relation
            </Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  );
};
