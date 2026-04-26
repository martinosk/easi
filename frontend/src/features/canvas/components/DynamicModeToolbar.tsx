import { Button, Group, Modal, Stack, Text } from '@mantine/core';
import { useState } from 'react';

interface DynamicModeToolbarProps {
  dirty: boolean;
  isSaving: boolean;
  saveLabel: string;
  onSave: () => void;
  onDiscard: () => void;
}

export function DynamicModeToolbar({
  dirty,
  isSaving,
  saveLabel,
  onSave,
  onDiscard,
}: DynamicModeToolbarProps) {
  const [confirmOpen, setConfirmOpen] = useState(false);

  const handleCancelClick = () => {
    if (!dirty) return;
    setConfirmOpen(true);
  };

  const handleConfirmDiscard = () => {
    setConfirmOpen(false);
    onDiscard();
  };

  return (
    <>
      <Group gap="xs">
        <Button color="green" disabled={isSaving || !dirty} onClick={onSave}>
          {saveLabel}
        </Button>
        <Button variant="default" disabled={isSaving || !dirty} onClick={handleCancelClick}>
          Cancel
        </Button>
      </Group>

      <Modal opened={confirmOpen} onClose={() => setConfirmOpen(false)} title="Discard changes?" centered>
        <Stack>
          <Text size="sm">You have unsaved changes in this view. Discarding will revert to the last saved state.</Text>
          <Group justify="flex-end" gap="xs">
            <Button variant="default" onClick={() => setConfirmOpen(false)}>
              Keep editing
            </Button>
            <Button color="red" onClick={handleConfirmDiscard}>
              Discard changes
            </Button>
          </Group>
        </Stack>
      </Modal>
    </>
  );
}
