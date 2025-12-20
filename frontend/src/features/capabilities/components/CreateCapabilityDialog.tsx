import React, { useState, useEffect } from 'react';
import { Modal, TextInput, Textarea, Select, Button, Group, Stack, Alert } from '@mantine/core';
import { useCreateCapability, useUpdateCapabilityMetadata } from '../hooks/useCapabilities';
import { useMaturityLevels, useStatuses } from '../../../hooks/useMetadata';

interface CreateCapabilityDialogProps {
  isOpen: boolean;
  onClose: () => void;
}

const DEFAULT_MATURITY_LEVELS = ['Genesis', 'Custom Build', 'Product', 'Commodity'];
const DEFAULT_STATUSES = [
  { value: 'Active', displayName: 'Active', sortOrder: 1 },
  { value: 'Planned', displayName: 'Planned', sortOrder: 2 },
  { value: 'Deprecated', displayName: 'Deprecated', sortOrder: 3 },
];

interface FormState {
  name: string;
  description: string;
  status: string;
  maturityLevel: string;
}

interface FormErrors {
  name?: string;
  description?: string;
}

const validateForm = (form: FormState): FormErrors => {
  const errors: FormErrors = {};

  if (!form.name.trim()) {
    errors.name = 'Name is required';
  } else if (form.name.trim().length > 200) {
    errors.name = 'Name must be 200 characters or less';
  }

  if (form.description.length > 1000) {
    errors.description = 'Description must be 1000 characters or less';
  }

  return errors;
};

export const CreateCapabilityDialog: React.FC<CreateCapabilityDialogProps> = ({
  isOpen,
  onClose,
}) => {
  const [form, setForm] = useState<FormState>({
    name: '',
    description: '',
    status: 'Active',
    maturityLevel: '',
  });
  const [errors, setErrors] = useState<FormErrors>({});
  const [backendError, setBackendError] = useState<string | null>(null);

  const { data: maturityLevelsData, isLoading: isLoadingMaturityLevels } = useMaturityLevels();
  const { data: statusesData, isLoading: isLoadingStatuses } = useStatuses();
  const createCapabilityMutation = useCreateCapability();
  const updateMetadataMutation = useUpdateCapabilityMetadata();

  const maturityLevels = maturityLevelsData ?? DEFAULT_MATURITY_LEVELS;
  const statuses = statusesData ?? DEFAULT_STATUSES;

  useEffect(() => {
    if (isOpen && maturityLevels.length > 0 && !form.maturityLevel) {
      setForm((prev) => ({ ...prev, maturityLevel: maturityLevels[0] }));
    }
  }, [isOpen, maturityLevels, form.maturityLevel]);

  const resetForm = () => {
    const defaultMaturity = maturityLevels.length > 0 ? maturityLevels[0] : DEFAULT_MATURITY_LEVELS[0];
    setForm({
      name: '',
      description: '',
      status: 'Active',
      maturityLevel: defaultMaturity,
    });
    setErrors({});
    setBackendError(null);
  };

  const handleClose = () => {
    resetForm();
    onClose();
  };

  const handleFieldChange = (
    field: keyof FormState,
    value: string
  ) => {
    setForm((prev) => ({ ...prev, [field]: value }));
    if (errors[field as keyof FormErrors]) {
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

    try {
      const capability = await createCapabilityMutation.mutateAsync({
        name: form.name.trim(),
        description: form.description.trim() || undefined,
        level: 'L1',
      });

      await updateMetadataMutation.mutateAsync({
        id: capability.id,
        request: {
          status: form.status,
          maturityLevel: form.maturityLevel,
        },
      });

      handleClose();
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'Failed to create capability');
    }
  };

  const statusOptions = [...statuses]
    .sort((a, b) => a.sortOrder - b.sortOrder)
    .map((s) => ({
      value: s.value,
      label: s.displayName,
    }));

  const isCreating = createCapabilityMutation.isPending || updateMetadataMutation.isPending;

  return (
    <Modal
      opened={isOpen}
      onClose={handleClose}
      title="Create Capability"
      centered
      data-testid="create-capability-dialog"
    >
      <form onSubmit={handleSubmit}>
        <Stack gap="md">
          <TextInput
            label="Name"
            placeholder="Enter capability name"
            value={form.name}
            onChange={(e) => handleFieldChange('name', e.currentTarget.value)}
            required
            withAsterisk
            autoFocus
            disabled={isCreating}
            error={errors.name}
            data-testid="capability-name-input"
          />

          <Textarea
            label="Description"
            placeholder="Enter capability description (optional)"
            value={form.description}
            onChange={(e) => handleFieldChange('description', e.currentTarget.value)}
            rows={3}
            disabled={isCreating}
            error={errors.description}
            data-testid="capability-description-input"
          />

          <Select
            label="Status"
            value={form.status}
            onChange={(value) => handleFieldChange('status', value || 'Active')}
            data={isLoadingStatuses ? [] : statusOptions}
            disabled={isCreating || isLoadingStatuses}
            data-testid="capability-status-select"
          />

          <Select
            label="Maturity Level"
            value={form.maturityLevel}
            onChange={(value) => handleFieldChange('maturityLevel', value || '')}
            data={isLoadingMaturityLevels ? [] : maturityLevels}
            disabled={isCreating || isLoadingMaturityLevels}
            data-testid="capability-maturity-select"
          />

          {backendError && (
            <Alert color="red" data-testid="create-capability-error">
              {backendError}
            </Alert>
          )}

          <Group justify="flex-end" gap="sm">
            <Button
              variant="default"
              onClick={handleClose}
              disabled={isCreating}
              data-testid="create-capability-cancel"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              loading={isCreating}
              disabled={isLoadingMaturityLevels || isLoadingStatuses || !form.name.trim()}
              data-testid="create-capability-submit"
            >
              Create
            </Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  );
};
