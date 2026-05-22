import React from 'react';
import { Button, Group, Modal, Stack, TextInput } from '@mantine/core';

interface CreateViewDialogProps {
  isOpen: boolean;
  viewName: string;
  onViewNameChange: (name: string) => void;
  onClose: () => void;
  onCreate: () => void;
}

export const CreateViewDialog: React.FC<CreateViewDialogProps> = ({
  isOpen,
  viewName,
  onViewNameChange,
  onClose,
  onCreate,
}) => {
  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') onCreate();
  };

  return (
    <Modal opened={isOpen} onClose={onClose} title="Create New View" centered>
      <Stack gap="md">
        <TextInput
          placeholder="View name"
          value={viewName}
          onChange={(e) => onViewNameChange(e.currentTarget.value)}
          onKeyDown={handleKeyDown}
          data-autofocus
        />
        <Group justify="flex-end" gap="sm">
          <Button variant="default" onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={onCreate}>Create</Button>
        </Group>
      </Stack>
    </Modal>
  );
};
