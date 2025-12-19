import React, { useState } from 'react';
import { Modal, TextInput, Button, Group, Stack, Alert } from '@mantine/core';
import { useAppStore } from '../../../store/appStore';

interface AddExpertDialogProps {
  isOpen: boolean;
  onClose: () => void;
  capabilityId: string;
}

interface FormState {
  name: string;
  role: string;
  contact: string;
}

interface FormErrors {
  name?: string;
  role?: string;
  contact?: string;
}

const validateForm = (form: FormState): FormErrors => {
  const errors: FormErrors = {};

  if (!form.name.trim()) {
    errors.name = 'Name is required';
  }

  if (!form.role.trim()) {
    errors.role = 'Role is required';
  }

  if (!form.contact.trim()) {
    errors.contact = 'Contact is required';
  }

  return errors;
};

export const AddExpertDialog: React.FC<AddExpertDialogProps> = ({
  isOpen,
  onClose,
  capabilityId,
}) => {
  const [form, setForm] = useState<FormState>({
    name: '',
    role: '',
    contact: '',
  });
  const [errors, setErrors] = useState<FormErrors>({});
  const [isAdding, setIsAdding] = useState(false);
  const [backendError, setBackendError] = useState<string | null>(null);

  const addCapabilityExpert = useAppStore((state) => state.addCapabilityExpert);

  const resetForm = () => {
    setForm({
      name: '',
      role: '',
      contact: '',
    });
    setErrors({});
    setBackendError(null);
  };

  const handleClose = () => {
    resetForm();
    onClose();
  };

  const handleFieldChange = (field: keyof FormState, value: string) => {
    setForm((prev) => ({ ...prev, [field]: value }));
    if (errors[field]) {
      setErrors((prev) => ({ ...prev, [field]: undefined }));
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
      await addCapabilityExpert(capabilityId as import('../../../api/types').CapabilityId, {
        expertName: form.name.trim(),
        expertRole: form.role.trim(),
        contactInfo: form.contact.trim(),
      });

      handleClose();
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'Failed to add expert');
    } finally {
      setIsAdding(false);
    }
  };

  return (
    <Modal
      opened={isOpen}
      onClose={handleClose}
      title="Add Expert"
      centered
      data-testid="add-expert-dialog"
    >
      <form onSubmit={handleSubmit}>
        <Stack gap="md">
          <TextInput
            label="Name"
            placeholder="Enter expert name"
            value={form.name}
            onChange={(e) => handleFieldChange('name', e.currentTarget.value)}
            required
            withAsterisk
            autoFocus
            disabled={isAdding}
            error={errors.name}
            data-testid="expert-name-input"
          />

          <TextInput
            label="Role"
            placeholder="Enter expert role"
            value={form.role}
            onChange={(e) => handleFieldChange('role', e.currentTarget.value)}
            required
            withAsterisk
            disabled={isAdding}
            error={errors.role}
            data-testid="expert-role-input"
          />

          <TextInput
            label="Contact"
            placeholder="Enter contact information"
            value={form.contact}
            onChange={(e) => handleFieldChange('contact', e.currentTarget.value)}
            required
            withAsterisk
            disabled={isAdding}
            error={errors.contact}
            data-testid="expert-contact-input"
          />

          {backendError && (
            <Alert color="red" data-testid="add-expert-error">
              {backendError}
            </Alert>
          )}

          <Group justify="flex-end" gap="sm">
            <Button
              variant="default"
              onClick={handleClose}
              disabled={isAdding}
              data-testid="add-expert-cancel"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              loading={isAdding}
              disabled={!form.name.trim() || !form.role.trim() || !form.contact.trim()}
              data-testid="add-expert-submit"
            >
              Add
            </Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  );
};
