import { Button, Group, Modal, Stack, Textarea, TextInput } from '@mantine/core';

interface StageFormOverlayProps {
  isEditing: boolean;
  formData: { name: string; description: string };
  onFormDataChange: (data: { name: string; description: string }) => void;
  onSubmit: () => void;
  onCancel: () => void;
}

export function StageFormOverlay({ isEditing, formData, onFormDataChange, onSubmit, onCancel }: StageFormOverlayProps) {
  return (
    <Modal opened onClose={onCancel} title={isEditing ? 'Edit Stage' : 'Add Stage'} centered data-testid="stage-form">
      <Stack gap="md">
        <TextInput
          id="stage-name"
          label="Name"
          value={formData.name}
          onChange={(e) => onFormDataChange({ ...formData, name: e.currentTarget.value })}
          placeholder="e.g. Discovery"
          maxLength={100}
          data-autofocus
        />
        <Textarea
          id="stage-description"
          label="Description"
          value={formData.description}
          onChange={(e) => onFormDataChange({ ...formData, description: e.currentTarget.value })}
          placeholder="Optional description..."
          maxLength={500}
          rows={3}
        />
        <Group justify="flex-end" gap="sm">
          <Button variant="default" onClick={onCancel}>
            Cancel
          </Button>
          <Button onClick={onSubmit} disabled={!formData.name.trim()}>
            {isEditing ? 'Save' : 'Add'}
          </Button>
        </Group>
      </Stack>
    </Modal>
  );
}
