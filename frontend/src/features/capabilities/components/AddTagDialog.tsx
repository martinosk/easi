import React, { useState } from 'react';
import { Modal, TextInput, Button, Group, Stack, Alert } from '@mantine/core';
import { useAppStore } from '../../../store/appStore';

interface AddTagDialogProps {
  isOpen: boolean;
  onClose: () => void;
  capabilityId: string;
}

interface FormState {
  tag: string;
}

interface FormErrors {
  tag?: string;
}

const validateForm = (form: FormState): FormErrors => {
  const errors: FormErrors = {};

  if (!form.tag.trim()) {
    errors.tag = 'Tag name is required';
  }

  return errors;
};

export const AddTagDialog: React.FC<AddTagDialogProps> = ({
  isOpen,
  onClose,
  capabilityId,
}) => {
  const [form, setForm] = useState<FormState>({
    tag: '',
  });
  const [errors, setErrors] = useState<FormErrors>({});
  const [isAdding, setIsAdding] = useState(false);
  const [backendError, setBackendError] = useState<string | null>(null);

  const addCapabilityTag = useAppStore((state) => state.addCapabilityTag);

  const resetForm = () => {
    setForm({
      tag: '',
    });
    setErrors({});
    setBackendError(null);
  };

  const handleClose = () => {
    resetForm();
    onClose();
  };

  const handleFieldChange = (value: string) => {
    setForm({ tag: value });
    if (errors.tag) {
      setErrors({});
    }
    if (backendError) {
      setBackendError(null);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setBackendError(null);

    const validationErrors = validateForm(form);
    if (Object.keys(validationErrors).length > 0) {
      setErrors(validationErrors);
      return;
    }

    setIsAdding(true);

    try {
      await addCapabilityTag(capabilityId as import('../../../api/types').CapabilityId, form.tag.trim());

      handleClose();
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'Failed to add tag');
    } finally {
      setIsAdding(false);
    }
  };

  return (
    <Modal
      opened={isOpen}
      onClose={handleClose}
      title="Add Tag"
      centered
      data-testid="add-tag-dialog"
    >
      <form onSubmit={handleSubmit}>
        <Stack gap="md">
          <TextInput
            label="Tag Name"
            placeholder="Enter tag name"
            value={form.tag}
            onChange={(e) => handleFieldChange(e.currentTarget.value)}
            required
            withAsterisk
            autoFocus
            disabled={isAdding}
            error={errors.tag}
            data-testid="tag-name-input"
          />

          {backendError && (
            <Alert color="red" data-testid="add-tag-error">
              {backendError}
            </Alert>
          )}

          <Group justify="flex-end" gap="sm">
            <Button
              variant="default"
              onClick={handleClose}
              disabled={isAdding}
              data-testid="add-tag-cancel"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              loading={isAdding}
              disabled={!form.tag.trim()}
              data-testid="add-tag-submit"
            >
              Add
            </Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  );
};
