import { Alert, Button, Group, List, Modal, ScrollArea, Stack, Text } from '@mantine/core';
import { useEffect } from 'react';

interface ConfirmationDialogProps {
  title: string;
  message: string;
  itemName?: string;
  itemNames?: string[];
  confirmText?: string;
  cancelText?: string;
  onConfirm: () => void;
  onCancel: () => void;
  isLoading?: boolean;
  error?: string | null;
}

function useEnterConfirms(onConfirm: () => void, isLoading: boolean) {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Enter' && !isLoading) onConfirm();
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [onConfirm, isLoading]);
}

export function ConfirmationDialog({
  title,
  message,
  itemName,
  itemNames,
  confirmText = 'Confirm',
  cancelText = 'Cancel',
  onConfirm,
  onCancel,
  isLoading = false,
  error = null,
}: ConfirmationDialogProps) {
  useEnterConfirms(onConfirm, isLoading);

  return (
    <Modal
      opened
      onClose={onCancel}
      title={title}
      centered
      closeOnClickOutside={!isLoading}
      closeOnEscape={!isLoading}
      data-testid="confirmation-dialog"
    >
      <Stack gap="md">
        <Text>{message}</Text>
        {itemName && <Text fw={600}>"{itemName}"</Text>}
        {itemNames && itemNames.length > 0 && (
          <ScrollArea.Autosize mah={150}>
            <List size="sm">
              {itemNames.map((name) => (
                <List.Item key={name}>{name}</List.Item>
              ))}
            </List>
          </ScrollArea.Autosize>
        )}
        {error ? (
          <Alert color="red" data-testid="confirmation-dialog-error">
            {error}
          </Alert>
        ) : (
          <Text c="orange" fw={600} size="sm">
            This action cannot be undone.
          </Text>
        )}
        <Group justify="flex-end" gap="sm">
          <Button variant="default" onClick={onCancel} disabled={isLoading}>
            {cancelText}
          </Button>
          <Button color="red" onClick={onConfirm} loading={isLoading}>
            {confirmText}
          </Button>
        </Group>
      </Stack>
    </Modal>
  );
}
