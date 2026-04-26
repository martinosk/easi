import { Button, Group, Modal, Stack, Text } from '@mantine/core';
import { useState } from 'react';

interface DynamicModeToolbarProps {
  enabled: boolean;
  dirty: boolean;
  isSaving: boolean;
  saveLabel: string;
  onEnable: () => void;
  onSave: () => void;
  onDiscard: () => void;
}

export function DynamicModeToolbar({
  enabled,
  dirty,
  isSaving,
  saveLabel,
  onEnable,
  onSave,
  onDiscard,
}: DynamicModeToolbarProps) {
  const [confirmOpen, setConfirmOpen] = useState(false);

  if (!enabled) {
    return (
      <Button variant="default" onClick={onEnable}>
        Dynamic mode
      </Button>
    );
  }

  const handleCancelClick = () => {
    if (dirty) {
      setConfirmOpen(true);
    } else {
      onDiscard();
    }
  };

  const handleConfirmDiscard = () => {
    setConfirmOpen(false);
    onDiscard();
  };

  return (
    <>
      <Group gap="xs">
        <Button color="green" disabled={isSaving} onClick={onSave}>
          {saveLabel}
        </Button>
        <Button variant="default" disabled={isSaving} onClick={handleCancelClick}>
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
